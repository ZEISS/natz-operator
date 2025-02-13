package controllers

import (
	"context"
	"math"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	natsv1alpha1 "github.com/zeiss/natz-operator/api/v1alpha1"
	"github.com/zeiss/natz-operator/pkg/status"
	"github.com/zeiss/pkg/cast"
	"github.com/zeiss/pkg/conv"
	"github.com/zeiss/pkg/k8s/finalizers"
	"github.com/zeiss/pkg/slices"
	"github.com/zeiss/pkg/utilx"
)

const (
	EventReasonActivationSynchronized EventReason = "ActivationSyncronized"
	EventReasonActivationFailed       EventReason = "ActivationFailed"
)

// NatsActivationReconciler ...
type NatsActivationReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// NewNatsActivationReconciler ...
func NewNatsActivationReconciler(mgr ctrl.Manager) *NatsActivationReconciler {
	return &NatsActivationReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(EventRecorderLabel),
	}
}

//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsactivations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsactivations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsactivations/finalizers,verbs=update

// Reconcile ...
// nolint:gocyclo
func (r *NatsActivationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	obj := &natsv1alpha1.NatsActivation{}
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if !obj.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, obj)
	}

	// get latest version of the account
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		return reconcile.Result{}, err
	}

	err := r.reconcileResources(ctx, obj)
	if err != nil {
		r.Recorder.Event(obj, corev1.EventTypeWarning, cast.String(EventReasonActivationFailed), "activation resources reconciliation failed")
		return r.ManageError(ctx, obj, err)
	}

	return r.ManageSuccess(ctx, obj)
}

func (r *NatsActivationReconciler) reconcileDelete(ctx context.Context, obj *natsv1alpha1.NatsActivation) (ctrl.Result, error) {
	obj.SetFinalizers(finalizers.RemoveFinalizer(obj, natsv1alpha1.FinalizerName))

	err := r.Update(ctx, obj)
	if err != nil && !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *NatsActivationReconciler) reconcileResources(ctx context.Context, obj *natsv1alpha1.NatsActivation) error {
	if err := r.reconcileStatus(ctx, obj); err != nil {
		return err
	}

	if err := r.reconcileActivation(ctx, obj); err != nil {
		return err
	}

	return nil
}

func (r *NatsActivationReconciler) reconcileActivation(ctx context.Context, obj *natsv1alpha1.NatsActivation) error {
	account := &natsv1alpha1.NatsAccount{}
	accountName := client.ObjectKey{
		Namespace: obj.Namespace,
		Name:      obj.Spec.AccountRef.Name,
	}

	if err := r.Get(ctx, accountName, account); err != nil {
		return err
	}

	skAccount := &natsv1alpha1.NatsAccount{}
	skAccountName := client.ObjectKey{
		Namespace: obj.Namespace,
		Name:      obj.Spec.AccountRef.Name,
	}

	if err := r.Get(ctx, skAccountName, skAccount); err != nil {
		return err
	}

	sk := &natsv1alpha1.NatsKey{}
	skName := client.ObjectKey{
		Namespace: obj.Namespace,
		Name:      obj.Spec.SignerKeyRef.Name,
	}

	if err := r.Get(ctx, skName, sk); err != nil {
		return err
	}

	skSecret := &corev1.Secret{}
	skSecretName := client.ObjectKey{
		Namespace: sk.Namespace,
		Name:      sk.Name,
	}

	if err := r.Get(ctx, skSecretName, skSecret); err != nil {
		return err
	}

	signerKp, err := nkeys.FromSeed(skSecret.Data[natsv1alpha1.SecretSeedDataKey])
	if err != nil {
		return err
	}

	token := jwt.NewActivationClaims(account.Status.PublicKey)
	token.Name = obj.Spec.Subject
	token.IssuerAccount = skAccount.Status.PublicKey

	token.NotBefore = obj.Spec.Start.Unix()
	token.Expires = obj.Spec.Expiry.Unix()

	token.Activation.ImportSubject = jwt.Subject(obj.Spec.Subject)
	token.Activation.ImportType = jwt.ExportType(obj.Spec.ExportType)

	t, err := token.Encode(signerKp)
	if err != nil {
		return err
	}

	obj.Status.JWT = t

	return nil
}

func (r *NatsActivationReconciler) reconcileStatus(_ context.Context, _ *natsv1alpha1.NatsActivation) error {
	return nil
}

// IsCreating ...
func (r *NatsActivationReconciler) IsCreating(obj *natsv1alpha1.NatsActivation) bool {
	return utilx.Or(obj.Status.Conditions == nil, slices.Len(obj.Status.Conditions) == 0)
}

// IsSynchronized ...
func (r *NatsActivationReconciler) IsSynchronized(obj *natsv1alpha1.NatsActivation) bool {
	return obj.Status.Phase == natsv1alpha1.ActivationSynchronized
}

// ManageError ...
func (r *NatsActivationReconciler) ManageError(ctx context.Context, obj *natsv1alpha1.NatsActivation, err error) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Error(err, "reconciling activation", "activation", obj.Name)

	if errors.IsNotFound(err) {
		return ctrl.Result{Requeue: true}, nil
	}

	status.SetNatzActivationCondition(obj, status.NewNatzActivationFailed(obj, err))
	obj.Status.Phase = natsv1alpha1.ActivationPhaseFailed

	if err := r.Client.Status().Update(ctx, obj); err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: time.Second}, err
	}

	r.Recorder.Event(obj, corev1.EventTypeWarning, conv.String(EventReasonActivationFailed), "account synchronization failed")

	var retryInterval time.Duration

	return reconcile.Result{
		RequeueAfter: time.Duration(math.Min(float64(retryInterval.Nanoseconds()*2), float64(time.Hour.Nanoseconds()*6))),
		Requeue:      true,
	}, nil
}

// ManageSuccess ...
func (r *NatsActivationReconciler) ManageSuccess(ctx context.Context, obj *natsv1alpha1.NatsActivation) (ctrl.Result, error) {
	obj.Status.Phase = natsv1alpha1.ActivationSynchronized
	obj.Status.LastUpdate = metav1.Now()
	status.SetNatzActivationCondition(obj, status.NewNatzActivationSynchronizedCondition(obj))

	err := r.Status().Update(ctx, obj)
	if err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: time.Second}, err
	}

	r.Recorder.Event(obj, corev1.EventTypeNormal, conv.String(EventReasonAccountSychronized), "account synchronized")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NatsActivationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&natsv1alpha1.NatsActivation{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
