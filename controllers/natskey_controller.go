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

	natsv1alpha1 "github.com/zeiss/natz-operator/api/v1alpha1"
	"github.com/zeiss/natz-operator/pkg/status"
	"github.com/zeiss/pkg/conv"
	"github.com/zeiss/pkg/slices"
	"github.com/zeiss/pkg/utilx"
	corev1 "k8s.io/api/core/v1"
)

const (
	EventReasonPrivateKeyFailed       EventReason = "PrivateKeyFailed"
	EventReasonPrivateKeySynchronized EventReason = "PrivateKeySynchronized"
)

// NatsPrivateKeyReconciler ...
type NatsPrivateKeyReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// NewNatsKeyReconciler ...
func NewNatsKeyReconciler(mgr ctrl.Manager) *NatsPrivateKeyReconciler {
	return &NatsPrivateKeyReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(EventRecorderLabel),
	}
}

//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsprivatekeys,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsprivatekeys/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsprivatekeys/finalizers,verbs=update
//+kubebuilder:rbac:groups=,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile ...
// nolint:gocyclo
func (r *NatsPrivateKeyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	sk := &natsv1alpha1.NatsKey{}
	if err := r.Get(ctx, req.NamespacedName, sk); err != nil {
		// Request object not found, could have been deleted after reconcile request.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !sk.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, sk)
	}

	return r.reconcileResources(ctx, sk)
}

func (r *NatsPrivateKeyReconciler) reconcilePrivateKey(ctx context.Context, obj *natsv1alpha1.NatsKey) error {
	if !controllerutil.ContainsFinalizer(obj, natsv1alpha1.FinalizerName) {
		controllerutil.AddFinalizer(obj, natsv1alpha1.FinalizerName)
		return r.Update(ctx, obj)
	}

	return nil
}

func (r *NatsPrivateKeyReconciler) reconcileResources(ctx context.Context, sk *natsv1alpha1.NatsKey) (ctrl.Result, error) {
	err := r.reconcileStatus(ctx, sk)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.reconcilePrivateKey(ctx, sk)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.reconcileSecret(ctx, sk)
	if err != nil {
		return ctrl.Result{}, err
	}

	return r.ManageSuccess(ctx, sk)
}

func (r *NatsPrivateKeyReconciler) reconcileStatus(ctx context.Context, sk *natsv1alpha1.NatsKey) error {
	phase := natsv1alpha1.KeyPhaseSynchronized

	if sk.Status.Phase != phase {
		sk.Status.Phase = phase

		return r.Status().Update(ctx, sk)
	}

	return nil
}

func (r *NatsPrivateKeyReconciler) reconcileSecret(ctx context.Context, sk *natsv1alpha1.NatsKey) error {
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
	secret.Type = natsv1alpha1.SecretNameKey
	secret.Annotations = map[string]string{
		natsv1alpha1.OwnerAnnotation: fmt.Sprintf("%s/%s", secret.Namespace, secret.Name),
	}

	keys, err := sk.Keys()
	if err != nil {
		return err
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
		r.Recorder.Event(sk, corev1.EventTypeWarning, conv.String(EventReasonPrivateKeyFailed), "secret creation failed")
		return err
	}

	if op == controllerutil.OperationResultCreated || op == controllerutil.OperationResultUpdated {
		r.Recorder.Event(sk, corev1.EventTypeNormal, conv.String(EventReasonPrivateKeySynchronized), "secret created or updated")
	}

	return nil
}

func (r *NatsPrivateKeyReconciler) reconcileDelete(ctx context.Context, sk *natsv1alpha1.NatsKey) (ctrl.Result, error) {
	// Get the associated secret
	secret := &corev1.Secret{}
	secretName := client.ObjectKey{
		Namespace: sk.Namespace,
		Name:      sk.Name,
	}

	if err := r.Get(ctx, secretName, secret); utilx.NotEmpty(client.IgnoreNotFound(err)) {
		return ctrl.Result{}, err
	}

	if sk.Spec.PreventDeletion && secret.ObjectMeta.DeletionTimestamp.IsZero() {
		if controllerutil.HasControllerReference(secret) {
			if err := controllerutil.RemoveControllerReference(sk, secret, r.Scheme); err != nil {
				return ctrl.Result{Requeue: true}, err
			}

			if err := r.Update(ctx, secret); err != nil {
				return ctrl.Result{Requeue: true}, err
			}
		}
	}

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
func (r *NatsPrivateKeyReconciler) IsCreating(obj *natsv1alpha1.NatsKey) bool {
	return utilx.Or(obj.Status.Conditions == nil, slices.Len(obj.Status.Conditions) == 0)
}

// IsSynchronized ...
func (r *NatsPrivateKeyReconciler) IsSynchronized(obj *natsv1alpha1.NatsKey) bool {
	return obj.Status.Phase == natsv1alpha1.KeyPhaseSynchronized
}

// ManageSuccess ...
func (r *NatsPrivateKeyReconciler) ManageSuccess(ctx context.Context, obj *natsv1alpha1.NatsKey) (ctrl.Result, error) {
	if r.IsSynchronized(obj) {
		return ctrl.Result{}, nil
	}

	status.SetNatzKeyCondition(obj, status.NewKeySychronizedCondition(obj))

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

	r.Recorder.Event(obj, corev1.EventTypeNormal, conv.String(EventReasonOperatorSynchronized), "key synchronized")

	return ctrl.Result{}, nil
}

// ManageError ...
func (r *NatsPrivateKeyReconciler) ManageError(ctx context.Context, obj *natsv1alpha1.NatsKey, err error) (ctrl.Result, error) {
	status.SetNatzKeyCondition(obj, status.NewKeyFailedCondition(obj, err))

	if err := r.Client.Status().Update(ctx, obj); err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: time.Second}, err
	}

	r.Recorder.Event(obj, corev1.EventTypeWarning, conv.String(EventReasonPrivateKeyFailed), "key synchronization failed")

	var retryInterval time.Duration

	return reconcile.Result{
		RequeueAfter: time.Duration(math.Min(float64(retryInterval.Nanoseconds()*2), float64(time.Hour.Nanoseconds()*6))),
		Requeue:      true,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NatsPrivateKeyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&natsv1alpha1.NatsKey{}).
		Owns(&corev1.Secret{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(r)
}
