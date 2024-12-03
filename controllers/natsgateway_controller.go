package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	natsv1alpha1 "github.com/zeiss/natz-operator/api/v1alpha1"
	"github.com/zeiss/pkg/k8s/finalizers"
)

const (
	EventReasonGatewaySucceeded = "GatewaySucceeded"
	EventReasonGatewayFailed    = "GatewayFailed"
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
	log := log.FromContext(ctx)

	gateway := &natsv1alpha1.NatsGateway{}
	if err := r.Get(ctx, req.NamespacedName, gateway); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if !gateway.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("processing deletion of gateway")

		if finalizers.HasFinalizer(gateway, natsv1alpha1.FinalizerName) {
			err := r.reconcileDelete(ctx, gateway)
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		// Delete
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil
}

func (r *NatsGatewayReconciler) reconcileDelete(ctx context.Context, gateway *natsv1alpha1.NatsGateway) error {
	log := log.FromContext(ctx)

	log.Info("reconcile delete gateway", "name", gateway.Name, "namespace", gateway.Namespace)

	gateway.SetFinalizers(finalizers.RemoveFinalizer(gateway, natsv1alpha1.FinalizerName))
	err := r.Update(ctx, gateway)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NatsGatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&natsv1alpha1.NatsAccount{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
