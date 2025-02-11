package controllers

import (
	"context"
	"math"
	"sync"
	"time"

	natsv1alpha1 "github.com/zeiss/natz-operator/api/v1alpha1"

	"github.com/nats-io/nats.go"
	"github.com/zeiss/pkg/conv"
	"github.com/zeiss/pkg/slices"
	"github.com/zeiss/pkg/utilx"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// NatsAccountServer takes NatsAccount and serves them to a nats server (cluster)
type NatsAccountServer struct {
	client.Client
	Scheme   *runtime.Scheme
	accounts sync.Map
	nc       *nats.Conn
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsaccounts,verbs=get;list;watch;create;update;patch;delete

// NewNatsAccountServer ...
func NewNatsAccountServer(mgr ctrl.Manager, nc *nats.Conn) *NatsAccountServer {
	return &NatsAccountServer{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		nc:       nc,
		Recorder: mgr.GetEventRecorderFor(EventRecorderLabel),
	}
}

// GetJWT ...
func (r *NatsAccountServer) GetJWT(publicKey string) (string, bool) {
	jwt, ok := r.accounts.Load(publicKey)
	if !ok {
		return "", false
	}

	return conv.String(jwt), true
}

// Reconcile ...
func (r *NatsAccountServer) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	account := &natsv1alpha1.NatsAccount{}

	if err := r.Get(ctx, req.NamespacedName, account); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if account.DeletionTimestamp != nil {
		return r.reconcileDelete(ctx, account)
	}

	err := r.reconcileAccount(ctx, account)
	if err != nil {
		return r.ManageError(ctx, account, err)
	}

	return r.ManageSuccess(ctx, account)
}

func (r *NatsAccountServer) reconcileDelete(_ context.Context, obj *natsv1alpha1.NatsAccount) (ctrl.Result, error) {
	r.accounts.Delete(obj.Status.PublicKey)

	return ctrl.Result{}, nil
}

func (r *NatsAccountServer) reconcileAccount(_ context.Context, obj *natsv1alpha1.NatsAccount) error {
	if !r.IsSynchronized(obj) {
		return nil
	}

	if utilx.And(utilx.NotEmpty(obj.Status.JWT), utilx.NotEmpty(obj.Status.PublicKey)) {
		r.accounts.Store(obj.Status.PublicKey, obj.Status.JWT)

		return r.nc.Publish("$SYS.REQ.CLAIMS.UPDATE", []byte(obj.Status.JWT))
	}

	return nil
}

// IsCreating ...
func (r *NatsAccountServer) IsCreating(obj *natsv1alpha1.NatsAccount) bool {
	return utilx.Or(obj.Status.Conditions == nil, slices.Len(obj.Status.Conditions) == 0)
}

// IsSynchronized ...
func (r *NatsAccountServer) IsSynchronized(obj *natsv1alpha1.NatsAccount) bool {
	return obj.Status.Phase == natsv1alpha1.AccountPhaseSynchronized
}

// ManageSuccess ...
func (r *NatsAccountServer) ManageSuccess(ctx context.Context, obj *natsv1alpha1.NatsAccount) (ctrl.Result, error) {
	if !r.IsSynchronized(obj) {
		return ctrl.Result{Requeue: true}, nil
	}

	if r.IsCreating(obj) {
		return ctrl.Result{Requeue: true}, nil
	}

	r.Recorder.Event(obj, corev1.EventTypeNormal, conv.String(EventReasonAccountAccessGranted), "account access granted")

	return ctrl.Result{}, nil
}

// ManageError ...
func (r *NatsAccountServer) ManageError(ctx context.Context, obj *natsv1alpha1.NatsAccount, err error) (ctrl.Result, error) {
	r.Recorder.Event(obj, corev1.EventTypeWarning, conv.String(EventReasonAccountAccessFailed), "account access failed")

	var retryInterval time.Duration

	return reconcile.Result{
		RequeueAfter: time.Duration(math.Min(float64(retryInterval.Nanoseconds()*2), float64(time.Hour.Nanoseconds()*6))),
		Requeue:      true,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NatsAccountServer) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&natsv1alpha1.NatsAccount{}).
		Complete(r)
}
