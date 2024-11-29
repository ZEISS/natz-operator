package controllers

import (
	"context"
	"fmt"

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
	corev1 "k8s.io/api/core/v1"
)

const (
	EventRecorderLabel = "natz-controller"
)

type EventReason string

const (
	EventReasonOperatorCreateFailed          EventReason = "OperatorCreateFailed"
	EventReasonOperatorUpdateFailed          EventReason = "OperatorUpdateFailed"
	EventReasonOperatorDeleteFailed          EventReason = "OperatorDeleteFailed"
	EventReasonOperatorSecretCreateSucceeded EventReason = "OperatorSecretCreateSucceeded"
	EventReasonOperatorSecretCreateFailed    EventReason = "OperatorSecretCreateFailed"
)

// NatsOperatorReconciler ...
type NatsOperatorReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// NewNatsOperatorReconciler ...
func NewNatsOperatorReconciler(mgr ctrl.Manager) *NatsOperatorReconciler {
	return &NatsOperatorReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(EventRecorderLabel),
	}
}

//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsoperators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsoperators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsoperators/finalizers,verbs=update
//+kubebuilder:rbac:groups=,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile ...
// nolint:gocyclo
func (r *NatsOperatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	operator := &natsv1alpha1.NatsOperator{}
	if err := r.Get(ctx, req.NamespacedName, operator); err != nil {
		// Request object not found, could have been deleted after reconcile request.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !operator.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("processing deletion of operator")

		if finalizers.HasFinalizer(operator, natsv1alpha1.FinalizerName) {
			err := r.reconcileDelete(ctx, operator)
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		// Delete
		return reconcile.Result{}, nil
	}

	// get latest version of the operator
	if err := r.Get(ctx, req.NamespacedName, operator); err != nil {
		log.Error(err, "operator not found", "operator", req.NamespacedName)

		return reconcile.Result{}, err
	}

	err := r.reconcileResources(ctx, operator)
	if err != nil {
		r.Recorder.Event(operator, corev1.EventTypeWarning, cast.String(EventReasonOperatorCreateFailed), "operator resources reconciliation failed")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *NatsOperatorReconciler) reconcileResources(ctx context.Context, operator *natsv1alpha1.NatsOperator) error {
	log := log.FromContext(ctx)

	err := r.reconcileStatus(ctx, operator)
	if err != nil {
		log.Error(err, "failed to reconcile status", "name", operator.Name, "namespace", operator.Namespace)
		return err
	}

	err = r.reconcileOperator(ctx, operator)
	if err != nil {
		log.Error(err, "failed to reconcile operator", "name", operator.Name, "namespace", operator.Namespace)
		return err
	}

	err = r.reconcileSecret(ctx, operator)
	if err != nil {
		log.Error(err, "failed to reconcile secret", "name", operator.Name, "namespace", operator.Namespace)
		return err
	}

	err = r.reconcileSystemAccount(ctx, operator)
	if err != nil {
		log.Error(err, "failed to reconcile system account", "name", operator.Name, "namespace", operator.Namespace)
		return err
	}

	err = r.reconcileServerConfig(ctx, operator)
	if err != nil {
		log.Error(err, "failed to reconcile server config", "name", operator.Name, "namespace", operator.Namespace)
		return err
	}

	return nil
}

func (r *NatsOperatorReconciler) reconcileSystemAccount(ctx context.Context, operator *natsv1alpha1.NatsOperator) error {
	log := log.FromContext(ctx)

	systemAccount := &natsv1alpha1.NatsAccount{}
	systemAccountName := client.ObjectKey{
		Namespace: operator.Namespace,
		Name:      fmt.Sprintf("%v-system", operator.Name),
	}

	err := r.Get(ctx, systemAccountName, systemAccount)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if !errors.IsNotFound(err) {
		return nil
	}

	systemAccount.Name = systemAccountName.Name
	systemAccount.Namespace = systemAccountName.Namespace

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, systemAccount, func() error {
		systemAccount.Spec = natsv1alpha1.NatsAccountSpec{
			AllowUserNamespaces: []string{
				operator.Namespace,
			},
			OperatorRef: corev1.ObjectReference{
				Namespace: operator.Namespace,
				Name:      operator.Name,
			},
			Exports: []natsv1alpha1.Export{
				{
					Name:                 "account-monitoring-services",
					Subject:              "$SYS.REQ.ACCOUNT.*.*",
					Type:                 natsv1alpha1.Service,
					ResponseType:         jwt.ResponseTypeStream,
					AccountTokenPosition: 4,
					Info: jwt.Info{
						Description: `Request account specific monitoring services for: SUBSZ, CONNZ, LEAFZ, JSZ and INFO`,
						InfoURL:     "https://docs.nats.io/nats-server/configuration/sys_accounts",
					},
				},
				{
					Name:                 "account-monitoring-streams",
					Subject:              "$SYS.ACCOUNT.*.>",
					Type:                 natsv1alpha1.Stream,
					AccountTokenPosition: 3,
					Info: jwt.Info{
						Description: `Account specific monitoring stream`,
						InfoURL:     "https://docs.nats.io/nats-server/configuration/sys_accounts",
					},
				},
			},
			Limits: natsv1alpha1.OperatorLimits{
				NatsLimits: jwt.NatsLimits{
					Subs:    -1,
					Payload: -1,
					Data:    -1,
				},
				AccountLimits: jwt.AccountLimits{
					Conn:            -1,
					Exports:         -1,
					WildcardExports: true,
					DisallowBearer:  true,
				},
			},
		}

		return nil
	})

	if op == controllerutil.OperationResultCreated || op == controllerutil.OperationResultUpdated {
		log.Info("system account created or updated", "operation", op)
	}

	return err
}

func (r *NatsOperatorReconciler) reconcileOperator(ctx context.Context, operator *natsv1alpha1.NatsOperator) error {
	log.FromContext(ctx)

	if controllerutil.AddFinalizer(operator, natsv1alpha1.FinalizerName) {
		if err := r.Update(ctx, operator); err != nil && !errors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func (r *NatsOperatorReconciler) reconcileServerConfig(ctx context.Context, operator *natsv1alpha1.NatsOperator) error {
	log := log.FromContext(ctx)

	log.Info("reconcile server config", "name", operator.Name, "namespace", operator.Namespace)

	serverConfig := &corev1.Secret{}
	serverConfigName := client.ObjectKey{
		Namespace: operator.Namespace,
		Name:      fmt.Sprintf("%v-server-config", operator.Name),
	}

	err := r.Get(ctx, serverConfigName, serverConfig)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if !errors.IsNotFound(err) {
		return err
	}

	systemAccount := &natsv1alpha1.NatsAccount{}
	systemAccountName := client.ObjectKey{
		Namespace: operator.Namespace,
		Name:      fmt.Sprintf("%v-system", operator.Name),
	}

	err = r.Get(ctx, systemAccountName, systemAccount)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	serverConfig.Namespace = operator.Namespace
	serverConfig.Name = serverConfigName.Name
	serverConfig.Type = "natz.zeiss.com/nats-configuration"

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, serverConfig, func() error {
		template := fmt.Sprintf(AUTH_CONFIG_TEMPLATE, operator.Status.JWT, systemAccount.Status.PublicKey, systemAccount.Status.PublicKey, systemAccount.Status.JWT)
		serverConfig.Data = map[string][]byte{
			OPERATOR_CONFIG_FILE: []byte(template),
		}

		return controllerutil.SetControllerReference(operator, serverConfig, r.Scheme)
	})

	if op == controllerutil.OperationResultCreated || op == controllerutil.OperationResultUpdated {
		log.Info("system account created or updated", "operation", op)
	}

	return err
}

func (r *NatsOperatorReconciler) reconcileStatus(ctx context.Context, operator *natsv1alpha1.NatsOperator) error {
	log := log.FromContext(ctx)

	log.Info("reconcile status", "name", operator.Name, "namespace", operator.Namespace)

	phase := utilx.IfElse(
		utilx.Empty(operator.Status.SecretName) && utilx.Empty(operator.Status.PublicKey) && utilx.Empty(operator.Status.JWT),
		natsv1alpha1.OperatorPhasePending,
		natsv1alpha1.OperatorPhaseSynchronized,
	)

	if operator.Status.Phase != phase {
		operator.Status.Phase = phase

		return r.Status().Update(ctx, operator)
	}

	return nil
}

func (r *NatsOperatorReconciler) reconcileDelete(ctx context.Context, s *natsv1alpha1.NatsOperator) error {
	log := log.FromContext(ctx)

	log.Info("reconcile delete operator", "name", s.Name, "namespace", s.Namespace)

	s.SetFinalizers(finalizers.RemoveFinalizer(s, natsv1alpha1.FinalizerName))
	err := r.Update(ctx, s)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	return nil
}

// nolint:gocyclo
func (r *NatsOperatorReconciler) reconcileSecret(ctx context.Context, operator *natsv1alpha1.NatsOperator) error {
	log := log.FromContext(ctx)

	log.Info("reconcile secret", "name", operator.Name, "namespace", operator.Namespace)

	secret := &corev1.Secret{}
	secretName := client.ObjectKey{
		Namespace: operator.Namespace,
		Name:      fmt.Sprintf("%v-secret", operator.Name),
	}

	err := r.Get(ctx, secretName, secret)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if !errors.IsNotFound(err) {
		return nil
	}

	keys, err := nkeys.CreateOperator()
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

	token := jwt.NewOperatorClaims(public)
	token.Operator.SigningKeys = operator.Spec.SigningKeys

	data := map[string][]byte{}
	data[OPERATOR_SEED_KEY] = seed
	data[OPERATOR_PUBLIC_KEY] = []byte(public)

	jwt, err := token.Encode(keys)
	if err != nil {
		return err
	}
	data[OPERATOR_JWT] = []byte(jwt)

	secret.Namespace = operator.Namespace
	secret.Name = secretName.Name
	secret.Type = "natz.zeiss.com/nats-operator"

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
		secret.Data = data

		return controllerutil.SetControllerReference(operator, secret, r.Scheme)
	})
	if err != nil {
		r.Recorder.Event(operator, corev1.EventTypeWarning, conv.String(EventReasonOperatorSecretCreateFailed), "secret creation failed")
		return err
	}

	if op == controllerutil.OperationResultCreated || op == controllerutil.OperationResultUpdated {
		r.Recorder.Event(operator, corev1.EventTypeNormal, conv.String(EventReasonOperatorSecretCreateSucceeded), "secret created or updated")

		log.Info("secret created or updated", "operation", op)
	}

	operator.Status.SecretName = secret.Name
	operator.Status.PublicKey = string(secret.Data[OPERATOR_PUBLIC_KEY])
	operator.Status.JWT = string(secret.Data[OPERATOR_JWT])

	if err := r.Status().Update(ctx, operator); err != nil {
		return err
	}

	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *NatsOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&natsv1alpha1.NatsOperator{}).
		Owns(&natsv1alpha1.NatsAccount{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

// nolint:gocyclo
func extractOrCreateKeys(secret *corev1.Secret, generator func() (nkeys.KeyPair, error)) (nkeys.KeyPair, bool, error) {
	var keys nkeys.KeyPair
	needsKeyUpdate := true

	if secret.Data != nil {
		parsedKeys, err := nkeys.FromSeed(secret.Data[OPERATOR_SEED_KEY])
		if err == nil {
			keys = parsedKeys
			needsKeyUpdate = false
		}
	}

	if keys == nil {
		// No keys present or failed to extract, create new key pair
		createdKeys, err := generator()
		if err != nil {
			return nil, false, err
		}
		keys = createdKeys
	}

	return keys, needsKeyUpdate, nil
}
