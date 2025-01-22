package controllers

import (
	"context"
	"math"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
	EventReasonAccountSucceed           = "AccountSucceed"
	EventReasonAccountFailed            = "AccountFailed"
	EventReasonAccountSychronized       = "AccountSychronized"
	EventReasonAccountSychronizedFailed = "AccountSychronizedFailed"
)

const (
	EventReasonAccountSecretCreateSucceeded EventReason = "AccountSecretCreateSucceeded"
	EventReasonAccountSecretCreateFailed    EventReason = "AccountSecretCreateFailed"
	EventReasonAccountAccessGranted         EventReason = "AccountAccessGranted"
	EventReasonAccountAccessDeleted         EventReason = "AccountAccessDeleted"
	EventReasonAccountAccessFailed          EventReason = "AccountAccessFailed"
)

// NatsAccountReconciler ...
type NatsAccountReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// NewNatsAccountReconciler ...
func NewNatsAccountReconciler(mgr ctrl.Manager) *NatsAccountReconciler {
	return &NatsAccountReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(EventRecorderLabel),
	}
}

//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsaccounts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsaccounts/finalizers,verbs=update

// Reconcile ...
// nolint:gocyclo
func (r *NatsAccountReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	account := &natsv1alpha1.NatsAccount{}
	if err := r.Get(ctx, req.NamespacedName, account); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if !account.ObjectMeta.DeletionTimestamp.IsZero() {
		if finalizers.HasFinalizer(account, natsv1alpha1.FinalizerName) {
			err := r.reconcileDelete(ctx, account)
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		// Delete
		return reconcile.Result{}, nil
	}

	// get latest version of the account
	if err := r.Get(ctx, req.NamespacedName, account); err != nil {
		return reconcile.Result{}, err
	}

	err := r.reconcileResources(ctx, account)
	if err != nil {
		r.Recorder.Event(account, corev1.EventTypeWarning, cast.String(EventReasonAccountSychronizedFailed), "account resources reconciliation failed")
		return r.ManageError(ctx, account, err)
	}

	return r.ManageSuccess(ctx, account)
}

func (r *NatsAccountReconciler) reconcileDelete(ctx context.Context, account *natsv1alpha1.NatsAccount) error {
	account.SetFinalizers(finalizers.RemoveFinalizer(account, natsv1alpha1.FinalizerName))
	err := r.Update(ctx, account)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	return nil
}

func (r *NatsAccountReconciler) reconcileResources(ctx context.Context, account *natsv1alpha1.NatsAccount) error {
	if err := r.reconcileStatus(ctx, account); err != nil {
		return err
	}

	if err := r.reconcileAccount(ctx, account); err != nil {
		return err
	}

	return nil
}

// nolint:gocyclo
func (r *NatsAccountReconciler) reconcileAccount(ctx context.Context, account *natsv1alpha1.NatsAccount) error {
	sk := &natsv1alpha1.NatsKey{}
	skName := client.ObjectKey{
		Namespace: account.Namespace,
		Name:      account.Spec.SignerKeyRef.Name,
	}

	if err := r.Get(ctx, skName, sk); errors.IsNotFound(err) {
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

	pk := &natsv1alpha1.NatsKey{}
	pkName := client.ObjectKey{
		Namespace: account.Namespace,
		Name:      account.Spec.PrivateKey.Name,
	}

	if err := r.Get(ctx, pkName, pk); err != nil {
		return err
	}

	pkSecret := &corev1.Secret{}
	pkSecretName := client.ObjectKey{
		Namespace: pk.Namespace,
		Name:      pk.Name,
	}

	if err := r.Get(ctx, pkSecretName, pkSecret); errors.IsNotFound(err) {
		return err
	}

	pkSigner, err := nkeys.FromSeed(pkSecret.Data[OPERATOR_SEED_KEY])
	if err != nil {
		return err
	}

	public, err := pkSigner.PublicKey()
	if err != nil {
		return err
	}

	signerKp, err := nkeys.FromSeed(skSecret.Data[OPERATOR_SEED_KEY])
	if err != nil {
		return err
	}

	token := jwt.NewAccountClaims(public)
	token.Name = account.Name
	token.Account = account.Spec.ToJWTAccount()

	for _, key := range account.Spec.SigningKeys {
		sk := &corev1.Secret{}
		skName := client.ObjectKey{
			Namespace: account.Namespace,
			Name:      key.Name,
		}

		if err := r.Get(ctx, skName, sk); err != nil {
			return err
		}

		skSigner, err := nkeys.FromSeed(sk.Data[OPERATOR_SEED_KEY])
		if err != nil {
			return err
		}

		pkSigner, err := skSigner.PublicKey()
		if err != nil {
			return err
		}

		token.SigningKeys.Add(pkSigner)
	}

	t, err := token.Encode(signerKp)
	if err != nil {
		return err
	}
	account.Status.JWT = t
	account.Status.PublicKey = public

	if !controllerutil.ContainsFinalizer(account, natsv1alpha1.FinalizerName) {
		controllerutil.AddFinalizer(account, natsv1alpha1.FinalizerName)
	}

	return nil
}

func (r *NatsAccountReconciler) reconcileStatus(ctx context.Context, account *natsv1alpha1.NatsAccount) error {
	phase := utilx.IfElse(
		utilx.Empty(account.Status.PublicKey) && utilx.Empty(account.Status.JWT),
		natsv1alpha1.AccountPhasePending,
		natsv1alpha1.AccountPhaseSynchronized,
	)

	if account.Status.Phase != phase {
		account.Status.Phase = phase

		return r.Status().Update(ctx, account)
	}

	return nil
}

// IsCreating ...
func (r *NatsAccountReconciler) IsCreating(obj *natsv1alpha1.NatsAccount) bool {
	return utilx.Or(obj.Status.Conditions == nil, slices.Len(obj.Status.Conditions) == 0)
}

// IsSynchronized ...
func (r *NatsAccountReconciler) IsSynchronized(obj *natsv1alpha1.NatsAccount) bool {
	return obj.Status.Phase == natsv1alpha1.AccountPhaseSynchronized
}

// ManageError ...
func (r *NatsAccountReconciler) ManageError(ctx context.Context, obj *natsv1alpha1.NatsAccount, err error) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Error(err, "reconciling account", "account", obj.Name)

	status.SetNatzAccountCondition(obj, status.NewAccountFailedCondition(obj, err))

	if err := r.Client.Status().Update(ctx, obj); err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: time.Second}, err
	}

	r.Recorder.Event(obj, corev1.EventTypeWarning, conv.String(EventReasonAccountSychronizedFailed), "account synchronization failed")

	var retryInterval time.Duration

	return reconcile.Result{
		RequeueAfter: time.Duration(math.Min(float64(retryInterval.Nanoseconds()*2), float64(time.Hour.Nanoseconds()*6))),
		Requeue:      true,
	}, nil
}

// ManageSuccess ...
func (r *NatsAccountReconciler) ManageSuccess(ctx context.Context, obj *natsv1alpha1.NatsAccount) (ctrl.Result, error) {
	if r.IsSynchronized(obj) {
		return ctrl.Result{}, nil
	}

	status.SetNatzAccountCondition(obj, status.NewAccountSychronizedCondition(obj))

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

	r.Recorder.Event(obj, corev1.EventTypeNormal, conv.String(EventReasonAccountSychronized), "account synchronized")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NatsAccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&natsv1alpha1.NatsAccount{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
