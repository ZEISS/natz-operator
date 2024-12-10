package controllers

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	natsv1alpha1 "github.com/zeiss/natz-operator/api/v1alpha1"
	"github.com/zeiss/natz-operator/pkg/status"
	"github.com/zeiss/pkg/conv"
	"github.com/zeiss/pkg/slices"
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
	EventReasonOperatorSynchronized          EventReason = "OperatorSynchronized"
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

	log.Info("reconcile operator", "name", req.Name, "namespace", req.Namespace)

	operator := &natsv1alpha1.NatsOperator{}
	if err := r.Get(ctx, req.NamespacedName, operator); err != nil {
		// Request object not found, could have been deleted after reconcile request.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !operator.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, operator)
	}

	return r.reconcileResources(ctx, operator)
}

func (r *NatsOperatorReconciler) reconcileResources(ctx context.Context, operator *natsv1alpha1.NatsOperator) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	err := r.reconcileStatus(ctx, operator)
	if err != nil {
		log.Error(err, "failed to reconcile status", "name", operator.Name, "namespace", operator.Namespace)
		return ctrl.Result{}, err
	}

	if err = r.reconcileOperator(ctx, operator); err != nil {
		return ctrl.Result{}, err
	}

	err = r.reconcileSystemAccount(ctx, operator)
	if err != nil {
		log.Error(err, "failed to reconcile system account", "name", operator.Name, "namespace", operator.Namespace)
		return ctrl.Result{}, err
	}

	res, err := r.reconcileServerConfig(ctx, operator)
	if err != nil {
		log.Error(err, "failed to reconcile server config", "name", operator.Name, "namespace", operator.Namespace)
		return res, err
	}

	return r.ManageSuccess(ctx, operator)
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
	if err != nil {
		return err
	}

	if op == controllerutil.OperationResultCreated || op == controllerutil.OperationResultUpdated {
		log.Info("system account created or updated", "operation", op)
	}

	return nil
}

func (r *NatsOperatorReconciler) reconcileOperator(ctx context.Context, obj *natsv1alpha1.NatsOperator) error {
	if !controllerutil.ContainsFinalizer(obj, natsv1alpha1.FinalizerName) {
		controllerutil.AddFinalizer(obj, natsv1alpha1.FinalizerName)
		return r.Update(ctx, obj)
	}

	return nil
}

func (r *NatsOperatorReconciler) reconcileServerConfig(ctx context.Context, operator *natsv1alpha1.NatsOperator) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info("reconcile server config", "name", operator.Name, "namespace", operator.Namespace)

	serverConfig := &corev1.Secret{}
	serverConfigName := client.ObjectKey{
		Namespace: operator.Namespace,
		Name:      fmt.Sprintf("%v-server-config", operator.Name),
	}

	err := r.Get(ctx, serverConfigName, serverConfig)
	if err != nil && !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	if !errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}

	systemAccount := &natsv1alpha1.NatsAccount{}
	systemAccountName := client.ObjectKey{
		Namespace: operator.Namespace,
		Name:      fmt.Sprintf("%v-system", operator.Name),
	}

	err = r.Get(ctx, systemAccountName, systemAccount)
	if err != nil && !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	if systemAccount.Status.Phase != natsv1alpha1.AccountPhaseSynchronized {
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: time.Second * 1,
		}, nil
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
	if err != nil {
		return ctrl.Result{}, err
	}

	if op == controllerutil.OperationResultCreated || op == controllerutil.OperationResultUpdated {
		log.Info("system account created or updated", "operation", op)
	}

	return ctrl.Result{}, nil
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

func (r *NatsOperatorReconciler) reconcileDelete(ctx context.Context, operator *natsv1alpha1.NatsOperator) (ctrl.Result, error) {
	// Remove our finalizer from the list.
	controllerutil.RemoveFinalizer(operator, natsv1alpha1.FinalizerName)

	if !operator.DeletionTimestamp.IsZero() {
		// Remove our finalizer from the list.
		controllerutil.RemoveFinalizer(operator, natsv1alpha1.FinalizerName)

		// Stop reconciliation as the object is being deleted.
		return ctrl.Result{}, r.Update(ctx, operator)
	}

	return ctrl.Result{Requeue: true}, nil
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
	// token.Operator.SigningKeys = operator.Spec.SigningKeys

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

	return nil
}

// IsCreating ...
func (r *NatsOperatorReconciler) IsCreating(obj *natsv1alpha1.NatsOperator) bool {
	return utilx.Or(obj.Status.Conditions == nil, slices.Len(obj.Status.Conditions) == 0)
}

// IsSynchronized ...
func (r *NatsOperatorReconciler) IsSynchronized(obj *natsv1alpha1.NatsOperator) bool {
	return obj.Status.Phase == natsv1alpha1.OperatorPhaseSynchronized
}

// ManageSuccess ...
func (r *NatsOperatorReconciler) ManageSuccess(ctx context.Context, obj *natsv1alpha1.NatsOperator) (ctrl.Result, error) {
	if r.IsSynchronized(obj) {
		return ctrl.Result{}, nil
	}

	status.SetNatzOperatorCondition(obj, status.NewOperatorSychronizedCondition(obj))

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

	r.Recorder.Event(obj, corev1.EventTypeNormal, conv.String(EventReasonOperatorSynchronized), "operator synchronized")

	return ctrl.Result{}, nil
}

// func (r *NatsOperatorReconciler) waitForSecretSync(obj *corev1.Secret) wait.ConditionWithContextFunc {
// 	return func(ctx context.Context) (bool, error) {
// 		newObj := &corev1.Secret{}
// 		if err := r.Get(ctx, client.ObjectKeyFromObject(obj), newObj); err != nil {
// 			if errors.IsNotFound(err) {
// 				return true, nil
// 			}

// 			return false, err
// 		}

// 		return equality.Semantic.DeepEqual(obj.Data, newObj.Data), nil
// 	}
// }

// SetupWithManager sets up the controller with the Manager.
func (r *NatsOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&natsv1alpha1.NatsOperator{}).
		Owns(&natsv1alpha1.NatsAccount{}).
		Owns(&corev1.Secret{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(r)
}
