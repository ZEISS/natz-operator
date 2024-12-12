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

	natsv1alpha1 "github.com/zeiss/natz-operator/api/v1alpha1"
	"github.com/zeiss/natz-operator/pkg/status"
	"github.com/zeiss/pkg/conv"
	"github.com/zeiss/pkg/slices"
	"github.com/zeiss/pkg/utilx"
)

const (
	EventReasonConfigSynchronizeFailed EventReason = "ConfigSynchronizeFailed"
	EventReasonConfigSynchronized      EventReason = "configSynchronized"
)

// NatsConfigReconciler reconciles a Natsconfig object
type NatsConfigReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// NewNatsConfigReconciler ...
func NewNatsConfigReconciler(mgr ctrl.Manager) *NatsConfigReconciler {
	return &NatsConfigReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(EventRecorderLabel),
	}
}

//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsconfig,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsconfig/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsconfig/finalizers,verbs=update

// Reconcile ...
func (r *NatsConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	config := &natsv1alpha1.NatsConfig{}
	if err := r.Get(ctx, req.NamespacedName, config); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if !config.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, config)
	}

	// get latest version of the config
	if err := r.Get(ctx, req.NamespacedName, config); err != nil {
		return reconcile.Result{}, err
	}

	return r.reconcileResources(ctx, config)
}

func (r *NatsConfigReconciler) reconcileDelete(ctx context.Context, obj *natsv1alpha1.NatsConfig) (ctrl.Result, error) {
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

func (r *NatsConfigReconciler) reconcileResources(ctx context.Context, config *natsv1alpha1.NatsConfig) (ctrl.Result, error) {
	if err := r.reconcileStatus(ctx, config); err != nil {
		return r.ManageError(ctx, config, err)
	}

	// if err := r.reconcileconfig(ctx, config); err != nil {
	// 	return r.ManageError(ctx, config, err)
	// }

	return r.ManageSuccess(ctx, config)
}

func (r *NatsConfigReconciler) reconcileConfig(ctx context.Context, config *natsv1alpha1.NatsConfig) error {
	// if !controllerutil.ContainsFinalizer(config, natsv1alpha1.FinalizerName) {
	// 	controllerutil.AddFinalizer(config, natsv1alpha1.FinalizerName)
	// }

	// if !controllerutil.HasControllerReference(config) {
	// 	if err := controllerutil.SetControllerReference(config, pk, r.Scheme); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func (r *NatsConfigReconciler) reconcileStatus(ctx context.Context, config *natsv1alpha1.NatsConfig) error {
	return nil
}

// IsCreating ...
func (r *NatsConfigReconciler) IsCreating(obj *natsv1alpha1.NatsConfig) bool {
	return utilx.Or(obj.Status.Conditions == nil, slices.Len(obj.Status.Conditions) == 0)
}

// IsSynchronized ...
func (r *NatsConfigReconciler) IsSynchronized(obj *natsv1alpha1.NatsConfig) bool {
	return obj.Status.Phase == natsv1alpha1.ConfigPhaseActive
}

// ManageError ...
func (r *NatsConfigReconciler) ManageError(ctx context.Context, obj *natsv1alpha1.NatsConfig, err error) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Error(err, "reconciliation failed", "config", obj)

	status.SetNatzConfigCondition(obj, status.NewNatzConfigFailedCondition(obj, err))

	if err := r.Client.Status().Update(ctx, obj); err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: time.Second}, err
	}

	r.Recorder.Event(obj, corev1.EventTypeWarning, conv.String(EventReasonConfigSynchronizeFailed), "config synchronization failed")

	var retryInterval time.Duration

	return reconcile.Result{
		RequeueAfter: time.Duration(math.Min(float64(retryInterval.Nanoseconds()*2), float64(time.Hour.Nanoseconds()*6))),
		Requeue:      true,
	}, nil
}

// ManageSuccess ...
func (r *NatsConfigReconciler) ManageSuccess(ctx context.Context, obj *natsv1alpha1.NatsConfig) (ctrl.Result, error) {
	if r.IsSynchronized(obj) {
		return ctrl.Result{}, nil
	}

	status.SetNatzConfigCondition(obj, status.NewNatzConfigSynchronizedCondition(obj))

	if r.IsCreating(obj) {
		return ctrl.Result{Requeue: true}, nil
	}

	if err := r.Client.Status().Update(ctx, obj); err != nil {
		return ctrl.Result{}, err
	}

	if !obj.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{Requeue: true}, nil
	}

	r.Recorder.Event(obj, corev1.EventTypeNormal, conv.String(EventReasonConfigSynchronized), "config synchronized")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NatsConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&natsv1alpha1.NatsConfig{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}