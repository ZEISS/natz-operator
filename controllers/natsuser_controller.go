package controllers

import (
	"context"
	"fmt"
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
	"github.com/zeiss/pkg/conv"
	"github.com/zeiss/pkg/slices"
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
	EventReasonUserSynchronized          EventReason = "UserSynchronized"
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
		return r.reconcileDelete(ctx, user)
	}

	// get latest version of the account
	if err := r.Get(ctx, req.NamespacedName, user); err != nil {
		log.Error(err, "account not found", "account", req.NamespacedName)

		return reconcile.Result{}, err
	}

	return r.reconcileResources(ctx, user)
}

func (r *NatsUserReconciler) reconcileDelete(ctx context.Context, obj *natsv1alpha1.NatsUser) (ctrl.Result, error) {
	// Remove our finalizer from the list.
	controllerutil.RemoveFinalizer(obj, natsv1alpha1.FinalizerName)

	if !obj.DeletionTimestamp.IsZero() {
		// Remove our finalizer from the list.
		controllerutil.RemoveFinalizer(obj, natsv1alpha1.FinalizerName)

		// Stop reconciliation as the object is being deleted.
		return ctrl.Result{}, r.Update(ctx, obj)
	}

	return ctrl.Result{Requeue: true}, nil
}

func (r *NatsUserReconciler) reconcileResources(ctx context.Context, user *natsv1alpha1.NatsUser) (ctrl.Result, error) {
	if err := r.reconcileStatus(ctx, user); err != nil {
		return r.ManageError(ctx, user, err)
	}

	if err := r.reconcileUser(ctx, user); err != nil {
		return r.ManageError(ctx, user, err)
	}

	if err := r.reconcileCredentials(ctx, user); err != nil {
		return r.ManageError(ctx, user, err)
	}

	return r.ManageSuccess(ctx, user)
}

func (r *NatsUserReconciler) reconcileCredentials(ctx context.Context, user *natsv1alpha1.NatsUser) error {
	secret := &corev1.Secret{}
	secretName := client.ObjectKey{
		Namespace: user.Namespace,
		Name:      fmt.Sprintf("%s-credentils", user.Name),
	}

	if err := r.Get(ctx, secretName, secret); !errors.IsNotFound(err) {
		return err
	}

	secret.Name = fmt.Sprintf("%s-credentials", user.Name)
	secret.Namespace = user.Namespace
	secret.Type = natsv1alpha1.SecretUserCredentialsName
	secret.Data = map[string][]byte{
		"user.jwt":   []byte(user.Status.JWT),
		"user.creds": []byte(fmt.Sprintf(ACCOUNT_TEMPLATE, user.Status.JWT, user.Spec.PrivateKey.Name)),
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
		return controllerutil.SetControllerReference(user, secret, r.Scheme)
	})
	if err != nil {
		return err
	}

	return nil
}

// nolint:gocyclo
func (r *NatsUserReconciler) reconcileUser(ctx context.Context, user *natsv1alpha1.NatsUser) error {
	sk := &natsv1alpha1.NatsSigningKey{}
	skName := client.ObjectKey{
		Namespace: user.Namespace,
		Name:      user.Spec.AccountSigningKey.Name,
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

	pk := &natsv1alpha1.NatsPrivateKey{}
	pkName := client.ObjectKey{
		Namespace: user.Namespace,
		Name:      user.Spec.PrivateKey.Name,
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

	token := jwt.NewUserClaims(public)
	token.User = user.Spec.ToNatsJWT()

	jwt, err := token.Encode(signerKp)
	if err != nil {
		return err
	}
	user.Status.JWT = jwt

	if !controllerutil.HasControllerReference(user) {
		if err := controllerutil.SetControllerReference(user, pk, r.Scheme); err != nil {
			return err
		}
	}

	return nil
}

func (r *NatsUserReconciler) reconcileStatus(ctx context.Context, user *natsv1alpha1.NatsUser) error {
	phase := utilx.IfElse(
		utilx.Empty(user.Status.PublicKey) && utilx.Empty(user.Status.JWT),
		natsv1alpha1.UserPhasePending,
		natsv1alpha1.UserPhaseSynchronized,
	)

	if user.Status.Phase != phase {
		user.Status.Phase = phase

		return r.Status().Update(ctx, user)
	}

	return nil
}

// IsCreating ...
func (r *NatsUserReconciler) IsCreating(obj *natsv1alpha1.NatsUser) bool {
	return utilx.Or(obj.Status.Conditions == nil, slices.Len(obj.Status.Conditions) == 0)
}

// IsSynchronized ...
func (r *NatsUserReconciler) IsSynchronized(obj *natsv1alpha1.NatsUser) bool {
	return obj.Status.Phase == natsv1alpha1.UserPhaseSynchronized
}

// ManageError ...
func (r *NatsUserReconciler) ManageError(ctx context.Context, obj *natsv1alpha1.NatsUser, err error) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Error(err, "reconciliation failed", "user", obj)

	status.SetNatzUserCondition(obj, status.NewUserFailedCondition(obj, err))

	if err := r.Client.Status().Update(ctx, obj); err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: time.Second}, err
	}

	r.Recorder.Event(obj, corev1.EventTypeWarning, conv.String(EventReasonUserSynchronizeFailed), "user synchronization failed")

	var retryInterval time.Duration

	return reconcile.Result{
		RequeueAfter: time.Duration(math.Min(float64(retryInterval.Nanoseconds()*2), float64(time.Hour.Nanoseconds()*6))),
		Requeue:      true,
	}, nil
}

// ManageSuccess ...
func (r *NatsUserReconciler) ManageSuccess(ctx context.Context, obj *natsv1alpha1.NatsUser) (ctrl.Result, error) {
	if r.IsSynchronized(obj) {
		return ctrl.Result{}, nil
	}

	status.SetNatzUserCondition(obj, status.NewUserSychronizedCondition(obj))

	if r.IsCreating(obj) {
		return ctrl.Result{Requeue: true}, nil
	}

	if err := r.Client.Status().Update(ctx, obj); err != nil {
		return ctrl.Result{}, err
	}

	if !obj.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{Requeue: true}, nil
	}

	r.Recorder.Event(obj, corev1.EventTypeNormal, conv.String(EventReasonUserSynchronized), "user synchronized")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NatsUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&natsv1alpha1.NatsUser{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
