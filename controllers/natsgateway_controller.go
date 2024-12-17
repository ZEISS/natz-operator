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
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	natsv1alpha1 "github.com/zeiss/natz-operator/api/v1alpha1"
	"github.com/zeiss/natz-operator/pkg/status"
	"github.com/zeiss/pkg/conv"
	"github.com/zeiss/pkg/k8s/finalizers"
	"github.com/zeiss/pkg/slices"
	"github.com/zeiss/pkg/utilx"
)

const (
	EventReasonGatewaySucceeded    = "GatewaySucceeded"
	EventReasonGatewaySynchronized = "GatewaySynchronized"
	EventReasonGatewayFailed       = "GatewayFailed"
)

// NatsGatewayReconciler ...
type NatsGatewayReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// NewNatsGatewayReconciler ...
func NewNatsGatewayReconciler(mgr ctrl.Manager) *NatsGatewayReconciler {
	return &NatsGatewayReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(EventRecorderLabel),
	}
}

//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsgateways,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsgateways/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsgateways/finalizers,verbs=update

// Reconcile ...
func (r *NatsGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	gateway := &natsv1alpha1.NatsGateway{}
	if err := r.Get(ctx, req.NamespacedName, gateway); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if !gateway.ObjectMeta.DeletionTimestamp.IsZero() {
		if finalizers.HasFinalizer(gateway, natsv1alpha1.FinalizerName) {
			err := r.reconcileDelete(ctx, gateway)
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		// Delete
		return reconcile.Result{}, nil
	}

	// get latest version of the gateway
	if err := r.Get(ctx, req.NamespacedName, gateway); err != nil {
		return reconcile.Result{}, err
	}

	return r.reconcileResources(ctx, req, gateway)
}

func (r *NatsGatewayReconciler) reconcileResources(ctx context.Context, req ctrl.Request, gateway *natsv1alpha1.NatsGateway) (ctrl.Result, error) {
	if err := r.reconcileGateway(ctx, req, gateway); err != nil {
		return r.ManageError(ctx, gateway, err)
	}

	if err := r.reconcileGateway(ctx, req, gateway); err != nil {
		return r.ManageError(ctx, gateway, err)
	}

	if err := r.reconcileUsername(ctx, gateway); err != nil {
		return r.ManageError(ctx, gateway, err)
	}

	if err := r.reconcilePassword(ctx, gateway); err != nil {
		return r.ManageError(ctx, gateway, err)
	}

	return r.ManageSuccess(ctx, gateway)
}

func (r *NatsGatewayReconciler) reconcileGateway(ctx context.Context, _ ctrl.Request, gateway *natsv1alpha1.NatsGateway) error {
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, gateway, func() error {
		controllerutil.AddFinalizer(gateway, natsv1alpha1.FinalizerName)

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *NatsGatewayReconciler) reconcilePassword(ctx context.Context, gateway *natsv1alpha1.NatsGateway) error {
	gatewayPwdSecret := &corev1.Secret{}
	gatewayPwdSecretName := client.ObjectKey{
		Namespace: gateway.Namespace,
		Name:      gateway.Spec.Password.SecretKeyRef.Name,
	}

	if err := r.Get(ctx, gatewayPwdSecretName, gatewayPwdSecret); err != nil {
		return err
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, gatewayPwdSecret, func() error {
		return controllerutil.SetControllerReference(gateway, gatewayPwdSecret, r.Scheme)
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *NatsGatewayReconciler) reconcileUsername(ctx context.Context, gateway *natsv1alpha1.NatsGateway) error {
	gatewayUserSecret := &corev1.Secret{}
	gatewayUserSecretName := client.ObjectKey{
		Namespace: gateway.Namespace,
		Name:      gateway.Spec.Username.SecretKeyRef.Name,
	}

	if err := r.Get(ctx, gatewayUserSecretName, gatewayUserSecret); err != nil {
		return err
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, gatewayUserSecret, func() error {
		return controllerutil.SetControllerReference(gateway, gatewayUserSecret, r.Scheme)
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *NatsGatewayReconciler) reconcileDelete(ctx context.Context, gateway *natsv1alpha1.NatsGateway) error {
	gateway.SetFinalizers(finalizers.RemoveFinalizer(gateway, natsv1alpha1.FinalizerName))
	err := r.Update(ctx, gateway)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	return nil
}

// ManageError ...
func (r *NatsGatewayReconciler) ManageError(ctx context.Context, obj *natsv1alpha1.NatsGateway, err error) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Error(err, "error reconciling gateway", "gateway", obj.Name)

	obj.Status.Phase = natsv1alpha1.GatewayPhaseFailed

	status.SetNatzGatewayCondition(obj, status.NewNatzGatewayFailedCondition(obj, err))

	if err := r.Client.Status().Update(ctx, obj); err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: time.Second}, err
	}

	r.Recorder.Event(obj, corev1.EventTypeWarning, conv.String(EventReasonGatewayFailed), "gateway synchronization failed")

	var retryInterval time.Duration

	return reconcile.Result{
		RequeueAfter: time.Duration(math.Min(float64(retryInterval.Nanoseconds()*2), float64(time.Hour.Nanoseconds()*6))),
		Requeue:      true,
	}, nil
}

// IsCreating ...
func (r *NatsGatewayReconciler) IsCreating(obj *natsv1alpha1.NatsGateway) bool {
	return utilx.Or(obj.Status.Conditions == nil, slices.Len(obj.Status.Conditions) == 0)
}

// IsSynchronized ...
func (r *NatsGatewayReconciler) IsSynchronized(obj *natsv1alpha1.NatsGateway) bool {
	return obj.Status.Phase == natsv1alpha1.GatewaySynchronized
}

// ManageSuccess ...
func (r *NatsGatewayReconciler) ManageSuccess(ctx context.Context, obj *natsv1alpha1.NatsGateway) (ctrl.Result, error) {
	if r.IsSynchronized(obj) {
		return ctrl.Result{}, nil
	}

	obj.Status.Phase = natsv1alpha1.GatewaySynchronized
	status.SetNatzGatewayCondition(obj, status.NewNatzGatewaySynchronizedCondition(obj))

	if r.IsCreating(obj) {
		return ctrl.Result{Requeue: true}, nil
	}

	if err := r.Client.Status().Update(ctx, obj); err != nil {
		return ctrl.Result{}, err
	}

	if !obj.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{Requeue: true}, nil
	}

	r.Recorder.Event(obj, corev1.EventTypeNormal, conv.String(EventReasonGatewaySynchronized), "gateway synchronized")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NatsGatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&natsv1alpha1.NatsGateway{}).
		Owns(&corev1.Secret{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(r)
}
