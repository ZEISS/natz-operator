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

const ACCOUNT_TEMPLATE = `-----BEGIN NATS USER JWT-----
%s
------END NATS USER JWT------

-----BEGIN USER NKEY SEED-----
%s
------END USER NKEY SEED------
`

const (
	EventReasonUserSecretCreateSucceeded EventReason = "UserSecretCreateSucceeded"
	EventReasonUserSecretCreateFailed    EventReason = "UserSecretCreateFailed"
	EventReasonUserSynchronizeFailed     EventReason = "UserSynchronizeFailed"
)

// NatsUserReconciler reconciles a NatsUser object
type NatsUserReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// NewNatsUserReconciler ...
func NewNatsUserReconciler(mgr ctrl.Manager) *NatsUserReconciler {
	return &NatsUserReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(EventRecorderLabel),
	}
}

//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsusers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsusers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsusers/finalizers,verbs=update

// Reconcile ...
// nolint:gocyclo
func (r *NatsUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	user := &natsv1alpha1.NatsUser{}
	if err := r.Get(ctx, req.NamespacedName, user); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if !user.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("processing deletion of user")

		if finalizers.HasFinalizer(user, natsv1alpha1.FinalizerName) {
			err := r.reconcileDelete(ctx, user)
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		// Delete
		return reconcile.Result{}, nil
	}

	// get latest version of the account
	if err := r.Get(ctx, req.NamespacedName, user); err != nil {
		log.Error(err, "account not found", "account", req.NamespacedName)

		return reconcile.Result{}, err
	}

	err := r.reconcileResources(ctx, req, user)
	if err != nil {
		r.Recorder.Event(user, corev1.EventTypeWarning, cast.String(EventReasonUserSynchronizeFailed), "user resources reconciliation failed")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *NatsUserReconciler) reconcileDelete(ctx context.Context, user *natsv1alpha1.NatsUser) error {
	log := log.FromContext(ctx)

	log.Info("reconcile delete user", "name", user.Name, "namespace", user.Namespace)

	user.SetFinalizers(finalizers.RemoveFinalizer(user, natsv1alpha1.FinalizerName))
	err := r.Update(ctx, user)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	return nil
}

func (r *NatsUserReconciler) reconcileResources(ctx context.Context, req ctrl.Request, user *natsv1alpha1.NatsUser) error {
	log := log.FromContext(ctx)

	log.Info("reconcile resources", "name", user.Name, "namespace", user.Namespace)

	if err := r.reconcileStatus(ctx, user); err != nil {
		log.Error(err, "failed to reconcile status", "name", user.Name, "namespace", user.Namespace)
		return err
	}

	if err := r.reconcileUser(ctx, req, user); err != nil {
		log.Error(err, "failed to reconcile user", "name", user.Name, "namespace", user.Namespace)
		return err
	}

	if err := r.reconcileSecret(ctx, user); err != nil {
		log.Error(err, "failed to reconcile secret", "name", user.Name, "namespace", user.Namespace)
		return err
	}

	return nil
}

func (r *NatsUserReconciler) reconcileUser(ctx context.Context, req ctrl.Request, user *natsv1alpha1.NatsUser) error {
	log := log.FromContext(ctx)

	issuer := &natsv1alpha1.NatsAccount{}
	issuerName := client.ObjectKey{
		Namespace: utilx.IfElse(utilx.Empty(user.Spec.AccountRef.Namespace), req.Namespace, user.Spec.AccountRef.Namespace),
		Name:      user.Spec.AccountRef.Name,
	}

	if err := r.Get(ctx, issuerName, issuer); errors.IsNotFound(err) {
		return err
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, user, func() error {
		controllerutil.AddFinalizer(user, natsv1alpha1.FinalizerName)

		return nil
	})
	if err != nil {
		return err
	}

	if op == controllerutil.OperationResultCreated || op == controllerutil.OperationResultUpdated {
		log.Info("user created or updated", "operation", op)
	}

	return nil
}

func (r *NatsUserReconciler) reconcileStatus(ctx context.Context, user *natsv1alpha1.NatsUser) error {
	log := log.FromContext(ctx)

	log.Info("reconcile status", "name", user.Name, "namespace", user.Namespace)

	phase := utilx.IfElse(
		utilx.Empty(user.Status.UserSecretName) && utilx.Empty(user.Status.PublicKey) && utilx.Empty(user.Status.JWT),
		natsv1alpha1.UserPhasePending,
		natsv1alpha1.UserPhaseSynchronized,
	)

	if user.Status.Phase != phase {
		user.Status.Phase = phase

		return r.Status().Update(ctx, user)
	}

	return nil
}

// nolint:gocyclo
func (r *NatsUserReconciler) reconcileSecret(ctx context.Context, user *natsv1alpha1.NatsUser) error {
	log := log.FromContext(ctx)

	log.Info("reconcile secret", "name", user.Name, "namespace", user.Namespace)

	issuer := &natsv1alpha1.NatsAccount{}
	issuerName := client.ObjectKey{
		Namespace: utilx.IfElse(utilx.Empty(user.Spec.AccountRef.Namespace), user.Namespace, user.Spec.AccountRef.Namespace),
		Name:      user.Spec.AccountRef.Name,
	}

	if err := r.Get(ctx, issuerName, issuer); errors.IsNotFound(err) {
		return err
	}

	signerSecret := &corev1.Secret{}
	signerSecretName := client.ObjectKey{
		Namespace: issuer.Namespace,
		Name:      issuer.Status.AccountSecretName,
	}

	if err := r.Get(ctx, signerSecretName, signerSecret); errors.IsNotFound(err) {
		return err
	}

	secret := &corev1.Secret{}
	secretName := client.ObjectKey{
		Namespace: user.Namespace,
		Name:      user.Name,
	}

	err := r.Get(ctx, secretName, secret)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if !errors.IsNotFound(err) {
		user.Status.JWT = string(secret.Data[OPERATOR_JWT])
		user.Status.PublicKey = string(secret.Data[OPERATOR_PUBLIC_KEY])
		user.Status.UserSecretName = secret.Name

		return r.Status().Update(ctx, user)
	}

	keys, err := nkeys.CreateUser()
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

	token := jwt.NewUserClaims(public)
	token.User = user.Spec.ToNatsJWT()

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

	secret.Namespace = user.Namespace
	secret.Name = secretName.Name
	secret.Type = "natz.zeiss.com/nats-user"

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
		secret.Data = data

		return controllerutil.SetControllerReference(user, secret, r.Scheme)
	})
	if err != nil {
		r.Recorder.Event(user, corev1.EventTypeWarning, conv.String(EventReasonOperatorSecretCreateFailed), "secret creation failed")
		return err
	}

	if op == controllerutil.OperationResultCreated || op == controllerutil.OperationResultUpdated {
		r.Recorder.Event(user, corev1.EventTypeNormal, conv.String(EventReasonUserSecretCreateSucceeded), "secret created or updated")

		log.Info("secret created or updated", "operation", op)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NatsUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&natsv1alpha1.NatsUser{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
