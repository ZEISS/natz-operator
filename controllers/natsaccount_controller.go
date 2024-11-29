package controllers

import (
	"context"

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
	"github.com/zeiss/pkg/cast"
	"github.com/zeiss/pkg/conv"
	"github.com/zeiss/pkg/k8s/finalizers"
	"github.com/zeiss/pkg/utilx"
)

const (
	EventReasonAccountSucceed = "AccountSucceed"
	EventReasonAccountFailed  = "AccountFailed"
)

const (
	EventReasonAccountSecretCreateSucceeded EventReason = "AccountSecretCreateSucceeded"
	EventReasonAccountSecretCreateFailed    EventReason = "AccountSecretCreateFailed"
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
	log := log.FromContext(ctx)

	account := &natsv1alpha1.NatsAccount{}
	if err := r.Get(ctx, req.NamespacedName, account); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if !account.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("processing deletion of account")

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
		log.Error(err, "account not found", "account", req.NamespacedName)

		return reconcile.Result{}, err
	}

	err := r.reconcileResources(ctx, req, account)
	if err != nil {
		r.Recorder.Event(account, corev1.EventTypeWarning, cast.String(EventReasonAccountFailed), "account resources reconciliation failed")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *NatsAccountReconciler) reconcileDelete(ctx context.Context, account *natsv1alpha1.NatsAccount) error {
	log := log.FromContext(ctx)

	log.Info("reconcile delete account", "name", account.Name, "namespace", account.Namespace)

	account.SetFinalizers(finalizers.RemoveFinalizer(account, natsv1alpha1.FinalizerName))
	err := r.Update(ctx, account)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	return nil
}

func (r *NatsAccountReconciler) reconcileResources(ctx context.Context, req ctrl.Request, account *natsv1alpha1.NatsAccount) error {
	log := log.FromContext(ctx)

	log.Info("reconcile resources", "name", account.Name, "namespace", account.Namespace)

	if err := r.reconcileStatus(ctx, account); err != nil {
		log.Error(err, "failed to reconcile status", "name", account.Name, "namespace", account.Namespace)
		return err
	}

	if err := r.reconcileAccount(ctx, req, account); err != nil {
		log.Error(err, "failed to reconcile account", "name", account.Name, "namespace", account.Namespace)
		return err
	}

	if err := r.reconcileSecret(ctx, account); err != nil {
		log.Error(err, "failed to reconcile secret", "name", account.Name, "namespace", account.Namespace)
		return err
	}

	return nil
}

func (r *NatsAccountReconciler) reconcileAccount(ctx context.Context, req ctrl.Request, account *natsv1alpha1.NatsAccount) error {
	log := log.FromContext(ctx)

	issuer := &natsv1alpha1.NatsOperator{}
	issuerName := client.ObjectKey{
		Namespace: req.Namespace,
		Name:      account.Spec.OperatorRef.Name,
	}

	if err := r.Get(ctx, issuerName, issuer); errors.IsNotFound(err) {
		return err
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, account, func() error {
		controllerutil.AddFinalizer(account, natsv1alpha1.FinalizerName)

		return controllerutil.SetControllerReference(issuer, account, r.Scheme)
	})
	if err != nil {
		return err
	}

	if op == controllerutil.OperationResultCreated || op == controllerutil.OperationResultUpdated {
		log.Info("account created or updated", "operation", op)
	}

	return nil
}

func (r *NatsAccountReconciler) reconcileStatus(ctx context.Context, account *natsv1alpha1.NatsAccount) error {
	log := log.FromContext(ctx)

	log.Info("reconcile status", "name", account.Name, "namespace", account.Namespace)

	phase := utilx.IfElse(
		utilx.Empty(account.Status.AccountSecretName) && utilx.Empty(account.Status.PublicKey) && utilx.Empty(account.Status.JWT),
		natsv1alpha1.AccountPhasePending,
		natsv1alpha1.AccountPhaseSynchronized,
	)

	if account.Status.Phase != phase {
		account.Status.Phase = phase

		return r.Status().Update(ctx, account)
	}

	return nil
}

// nolint:gocyclo
func (r *NatsAccountReconciler) reconcileSecret(ctx context.Context, account *natsv1alpha1.NatsAccount) error {
	log := log.FromContext(ctx)

	log.Info("reconcile secret", "name", account.Name, "namespace", account.Namespace)

	issuer := &natsv1alpha1.NatsOperator{}
	issuerName := client.ObjectKey{
		Namespace: account.Namespace,
		Name:      account.Spec.OperatorRef.Name,
	}

	if err := r.Get(ctx, issuerName, issuer); errors.IsNotFound(err) {
		return err
	}

	signerSecret := &corev1.Secret{}
	signerSecretName := client.ObjectKey{
		Namespace: issuer.Namespace,
		Name:      issuer.Status.SecretName,
	}

	if err := r.Get(ctx, signerSecretName, signerSecret); errors.IsNotFound(err) {
		return err
	}

	secret := &corev1.Secret{}
	secretName := client.ObjectKey{
		Namespace: account.Namespace,
		Name:      account.Name,
	}

	err := r.Get(ctx, secretName, secret)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if !errors.IsNotFound(err) {
		return nil
	}

	keys, err := nkeys.CreateAccount()
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

	token := jwt.NewAccountClaims(public)
	token.Account = account.Spec.ToJWTAccount()

	signerKp, err := nkeys.FromSeed(signerSecret.Data[OPERATOR_SEED_KEY])
	if err != nil {
		return err
	}

	data := map[string][]byte{}
	data[OPERATOR_SEED_KEY] = seed
	data[OPERATOR_PUBLIC_KEY] = []byte(public)

	jwt, err := token.Encode(signerKp)
	if err != nil {
		return err
	}
	data[OPERATOR_JWT] = []byte(jwt)

	secret.Namespace = account.Namespace
	secret.Name = secretName.Name
	secret.Type = "natz.zeiss.com/nats-account"

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
		secret.Data = data

		return controllerutil.SetControllerReference(account, secret, r.Scheme)
	})
	if err != nil {
		r.Recorder.Event(account, corev1.EventTypeWarning, conv.String(EventReasonOperatorSecretCreateFailed), "secret creation failed")
		return err
	}

	if op == controllerutil.OperationResultCreated || op == controllerutil.OperationResultUpdated {
		r.Recorder.Event(account, corev1.EventTypeNormal, conv.String(EventReasonAccountSecretCreateSucceeded), "secret created or updated")

		log.Info("secret created or updated", "operation", op)
	}

	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *NatsAccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&natsv1alpha1.NatsAccount{}).
		Owns(&corev1.Secret{}).
		Owns(&natsv1alpha1.NatsUser{}).
		Complete(r)
}
