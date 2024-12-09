package controllers

import (
	"context"
	"fmt"
	"math"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/nats-io/nkeys"
	"github.com/zeiss/natz-operator/api/v1alpha1"
	natsv1alpha1 "github.com/zeiss/natz-operator/api/v1alpha1"
	"github.com/zeiss/natz-operator/pkg/status"
	"github.com/zeiss/pkg/conv"
	"github.com/zeiss/pkg/slices"
	"github.com/zeiss/pkg/utilx"
	corev1 "k8s.io/api/core/v1"
)

const (
	EventReasonSigningKeyFailed       EventReason = "SigningKeyFailed"
	EventReasonSigningKeySynchronized EventReason = "SigningKeySynchronized"
)

// NatsSigningKeyReconciler ...
type NatsSigningKeyReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// NewNatsSigningKeyReconciler ...
func NewNatsSigningKeyReconciler(mgr ctrl.Manager) *NatsSigningKeyReconciler {
	return &NatsSigningKeyReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(EventRecorderLabel),
	}
}

//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natssigningkeys,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natssigningkeys/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natssigningkeys/finalizers,verbs=update
//+kubebuilder:rbac:groups=,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile ...
// nolint:gocyclo
func (r *NatsSigningKeyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	sk := &natsv1alpha1.NatsSigningKey{}
	if err := r.Get(ctx, req.NamespacedName, sk); err != nil {
		// Request object not found, could have been deleted after reconcile request.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !sk.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, sk)
	}

	return r.reconcileResources(ctx, sk)
}

func (r *NatsSigningKeyReconciler) reconcileSigningKey(ctx context.Context, obj *natsv1alpha1.NatsSigningKey) error {
	if !controllerutil.ContainsFinalizer(obj, natsv1alpha1.FinalizerName) {
		controllerutil.AddFinalizer(obj, natsv1alpha1.FinalizerName)
		return r.Update(ctx, obj)
	}

	return nil
}

func (r *NatsSigningKeyReconciler) reconcileResources(ctx context.Context, sk *natsv1alpha1.NatsSigningKey) (ctrl.Result, error) {
	err := r.reconcileStatus(ctx, sk)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.reconcileSigningKey(ctx, sk)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.reconcileSecret(ctx, sk)
	if err != nil {
		return ctrl.Result{}, err
	}

	return r.ManageSuccess(ctx, sk)
}

func (r *NatsSigningKeyReconciler) reconcileStatus(ctx context.Context, sk *natsv1alpha1.NatsSigningKey) error {
	phase := natsv1alpha1.SigningKeyPhaseSynchronized

	if sk.Status.Phase != phase {
		sk.Status.Phase = phase

		return r.Status().Update(ctx, sk)
	}

	return nil
}

func (r *NatsSigningKeyReconciler) reconcileSecret(ctx context.Context, sk *natsv1alpha1.NatsSigningKey) error {
	secret := &corev1.Secret{}
	secretName := client.ObjectKey{
		Namespace: sk.Namespace,
		Name:      sk.Name,
	}

	err := r.Get(ctx, secretName, secret)
	if !errors.IsNotFound(err) {
		return err
	}

	secret.Namespace = sk.Namespace
	secret.Name = sk.Name
	secret.Type = "natz.zeiss.com/nats-signing-key"
	secret.Annotations = map[string]string{
		v1alpha1.OwnerAnnotation: fmt.Sprintf("%s/%s", secret.Namespace, secret.Name),
	}

	var keys nkeys.KeyPair
	switch sk.Spec.Type {
	case "Operator":
		keys, err = nkeys.CreateOperator()
		if err != nil {
			return err
		}
	case "Account":
		keys, err = nkeys.CreateAccount()
		if err != nil {
			return err
		}
	case "User":
		keys, err = nkeys.CreateUser()
		if err != nil {
			return err
		}
	}

	seed, err := keys.Seed()
	if err != nil {
		return err
	}

	public, err := keys.PublicKey()
	if err != nil {
		return err
	}

	data := map[string][]byte{}
	data[OPERATOR_SEED_KEY] = seed
	data[OPERATOR_PUBLIC_KEY] = []byte(public)

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
		secret.Data = data

		return controllerutil.SetControllerReference(sk, secret, r.Scheme)
	})
	if err != nil {
		r.Recorder.Event(sk, corev1.EventTypeWarning, conv.String(EventReasonSigningKeyFailed), "secret creation failed")
		return err
	}

	if op == controllerutil.OperationResultCreated || op == controllerutil.OperationResultUpdated {
		r.Recorder.Event(sk, corev1.EventTypeNormal, conv.String(EventReasonSigningKeySynchronized), "secret created or updated")
	}

	return nil
}

func (r *NatsSigningKeyReconciler) reconcileDelete(ctx context.Context, sk *natsv1alpha1.NatsSigningKey) (ctrl.Result, error) {
	// Remove our finalizer from the list.
	controllerutil.RemoveFinalizer(sk, natsv1alpha1.FinalizerName)

	if !sk.DeletionTimestamp.IsZero() {
		// Remove our finalizer from the list.
		controllerutil.RemoveFinalizer(sk, natsv1alpha1.FinalizerName)

		// Stop reconciliation as the object is being deleted.
		return ctrl.Result{}, r.Update(ctx, sk)
	}

	return ctrl.Result{Requeue: true}, nil
}

// IsCreating ...
func (r *NatsSigningKeyReconciler) IsCreating(obj *natsv1alpha1.NatsSigningKey) bool {
	return utilx.Or(obj.Status.Conditions == nil, slices.Len(obj.Status.Conditions) == 0)
}

// IsSynchronized ...
func (r *NatsSigningKeyReconciler) IsSynchronized(obj *natsv1alpha1.NatsSigningKey) bool {
	return obj.Status.Phase == natsv1alpha1.SigningKeyPhaseSynchronized
}

// ManageSuccess ...
func (r *NatsSigningKeyReconciler) ManageSuccess(ctx context.Context, obj *natsv1alpha1.NatsSigningKey) (ctrl.Result, error) {
	if r.IsSynchronized(obj) {
		return ctrl.Result{}, nil
	}

	status.SetNatzSigningKeyCondition(obj, status.NewSigningKeySychronizedCondition(obj))

	if r.IsCreating(obj) {
		return ctrl.Result{Requeue: true}, nil
	}

	if err := r.Client.Status().Update(ctx, obj); err != nil {
		return ctrl.Result{}, err
	}

	if !obj.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{Requeue: true}, nil
	}

	if r.IsCreating(obj) {
		return ctrl.Result{Requeue: true}, nil
	}

	r.Recorder.Event(obj, corev1.EventTypeNormal, conv.String(EventReasonOperatorSynchronized), "signing key synchronized")

	return ctrl.Result{}, nil
}

// ManageError ...
func (r *NatsSigningKeyReconciler) ManageError(ctx context.Context, obj *natsv1alpha1.NatsSigningKey, err error) (ctrl.Result, error) {
	status.SetNatzSigningKeyCondition(obj, status.NewSigningKeyFailedCondition(obj, err))

	if err := r.Client.Status().Update(ctx, obj); err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: time.Second}, err
	}

	r.Recorder.Event(obj, corev1.EventTypeWarning, conv.String(EventReasonSigningKeyFailed), "secret synchronization failed")

	var retryInterval time.Duration

	return reconcile.Result{
		RequeueAfter: time.Duration(math.Min(float64(retryInterval.Nanoseconds()*2), float64(time.Hour.Nanoseconds()*6))),
		Requeue:      true,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NatsSigningKeyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&natsv1alpha1.NatsSigningKey{}).
		Owns(&corev1.Secret{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(r)
}
