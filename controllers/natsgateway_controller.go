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
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	natsv1alpha1 "github.com/zeiss/natz-operator/api/v1alpha1"
	"github.com/zeiss/pkg/cast"
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

	err := r.reconcileResources(ctx, req, gateway)
	if err != nil {
		r.Recorder.Event(gateway, corev1.EventTypeWarning, cast.String(EventReasonGatewayFailed), "gateway resources reconciliation failed")
		gateway.Status.Phase = natsv1alpha1.GatewayPhaseFailed
		return reconcile.Result{}, r.Status().Update(ctx, gateway)
	}

	return reconcile.Result{}, nil
}

func (r *NatsGatewayReconciler) reconcileResources(ctx context.Context, req ctrl.Request, gateway *natsv1alpha1.NatsGateway) error {
	if err := r.reconcileStatus(ctx, gateway); err != nil {
		return err
	}

	if err := r.reconcileGateway(ctx, req, gateway); err != nil {
		return err
	}

	if err := r.reconcileSecret(ctx, gateway); err != nil {
		return err
	}

	return nil
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

func (r *NatsGatewayReconciler) reconcileSecret(ctx context.Context, gateway *natsv1alpha1.NatsGateway) error {
	gatewaySecret := &corev1.Secret{}
	gatewaySecretName := client.ObjectKey{
		Namespace: gateway.Namespace,
		Name:      gateway.Spec.Password.SecretKeyRef.Name,
	}

	if err := r.Get(ctx, gatewaySecretName, gatewaySecret); errors.IsNotFound(err) {
		r.Recorder.Event(gateway, corev1.EventTypeWarning, cast.String(EventReasonGatewayFailed), "gateway secret not found")
		return err
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, gatewaySecret, func() error {
		return controllerutil.SetControllerReference(gateway, gatewaySecret, r.Scheme)
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *NatsGatewayReconciler) reconcileStatus(ctx context.Context, gateway *natsv1alpha1.NatsGateway) error {
	phase := natsv1alpha1.GatewayPhaseNone

	if gateway.Status.Phase != phase {
		gateway.Status.Phase = phase

		return r.Status().Update(ctx, gateway)
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

// SetupWithManager sets up the controller with the Manager.
func (r *NatsGatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&natsv1alpha1.NatsGateway{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
