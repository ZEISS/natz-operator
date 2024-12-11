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
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

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
	EventReasonOperatorSynchronizeFailed     EventReason = "OperatorSynchronizeFailed"
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

	err := r.reconcileResources(ctx, operator)
	if err != nil {
		return r.ManageError(ctx, operator, err)
	}

	return r.ManageSuccess(ctx, operator)
}

func (r *NatsOperatorReconciler) reconcileResources(ctx context.Context, operator *natsv1alpha1.NatsOperator) error {
	err := r.reconcileStatus(ctx, operator)
	if err != nil {
		return err
	}

	if err = r.reconcileOperator(ctx, operator); err != nil {
		return err
	}

	err = r.reconcileServerConfig(ctx, operator)
	if err != nil {
		return err
	}

	return err
}

func (r *NatsOperatorReconciler) reconcileOperator(ctx context.Context, obj *natsv1alpha1.NatsOperator) error {
	pk := &corev1.Secret{}
	pkName := client.ObjectKey{
		Namespace: obj.Namespace,
		Name:      obj.Spec.PrivateKey.Name,
	}

	if err := r.Get(ctx, pkName, pk); err != nil {
		return err
	}

	seed, ok := pk.Data[OPERATOR_SEED_KEY]
	if !ok {
		return fmt.Errorf("public key not found")
	}

	sk, err := nkeys.FromSeed(seed)
	if err != nil {
		return err
	}

	public, err := sk.PublicKey()
	if err != nil {
		return err
	}

	token := jwt.NewOperatorClaims(public)
	jwt, err := token.Encode(sk)
	if err != nil {
		return err
	}

	obj.Status.JWT = jwt
	obj.Status.PublicKey = public

	if !controllerutil.ContainsFinalizer(obj, natsv1alpha1.FinalizerName) {
		controllerutil.AddFinalizer(obj, natsv1alpha1.FinalizerName)
	}

	return nil
}

func (r *NatsOperatorReconciler) reconcileServerConfig(ctx context.Context, operator *natsv1alpha1.NatsOperator) error {
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
		return nil
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

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, serverConfig, func() error {
		template := fmt.Sprintf(AUTH_CONFIG_TEMPLATE, operator.Status.JWT, systemAccount.Status.PublicKey, systemAccount.Status.PublicKey, systemAccount.Status.JWT)
		serverConfig.Data = map[string][]byte{
			OPERATOR_CONFIG_FILE: []byte(template),
		}

		return controllerutil.SetControllerReference(operator, serverConfig, r.Scheme)
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *NatsOperatorReconciler) reconcileStatus(ctx context.Context, operator *natsv1alpha1.NatsOperator) error {
	phase := utilx.IfElse(
		utilx.Empty(operator.Status.PublicKey) && utilx.Empty(operator.Status.JWT),
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

// IsCreating ...
func (r *NatsOperatorReconciler) IsCreating(obj *natsv1alpha1.NatsOperator) bool {
	return utilx.Or(obj.Status.Conditions == nil, slices.Len(obj.Status.Conditions) == 0)
}

// IsSynchronized ...
func (r *NatsOperatorReconciler) IsSynchronized(obj *natsv1alpha1.NatsOperator) bool {
	return obj.Status.Phase == natsv1alpha1.OperatorPhaseSynchronized
}

// ManageError ...
func (r *NatsOperatorReconciler) ManageError(ctx context.Context, obj *natsv1alpha1.NatsOperator, err error) (ctrl.Result, error) {
	status.SetNatzOperatorCondition(obj, status.NewOperatorFailedCondition(obj, err))

	if err := r.Client.Status().Update(ctx, obj); err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: time.Second}, err
	}

	r.Recorder.Event(obj, corev1.EventTypeWarning, conv.String(EventReasonOperatorSynchronizeFailed), "operator synchronization failed")

	var retryInterval time.Duration

	return reconcile.Result{
		RequeueAfter: time.Duration(math.Min(float64(retryInterval.Nanoseconds()*2), float64(time.Hour.Nanoseconds()*6))),
		Requeue:      true,
	}, nil
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

// SetupWithManager sets up the controller with the Manager.
func (r *NatsOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&natsv1alpha1.NatsOperator{}).
		Owns(&natsv1alpha1.NatsAccount{}).
		Owns(&corev1.Secret{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(r)
}
