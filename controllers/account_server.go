package controllers

import (
	"context"

	natsv1alpha1 "github.com/zeiss/natz-operator/api/v1alpha1"

	"github.com/nats-io/nats.go"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// NatsAccountServer takes NatsAccount and serves them to a nats server (cluster)
type NatsAccountServer struct {
	client.Client
	Scheme   *runtime.Scheme
	accounts map[string]string
	nc       *nats.Conn
}

//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsaccounts,verbs=get;list;watch;create;update;patch;delete

// NewNatsAccountServer ...
func NewNatsAccountServer(mgr ctrl.Manager, nc *nats.Conn) *NatsAccountServer {
	return &NatsAccountServer{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		accounts: make(map[string]string),
		nc:       nc,
	}
}

// Reconcile ...
func (r *NatsAccountServer) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	account := &natsv1alpha1.NatsAccount{}
	if err := r.Get(ctx, req.NamespacedName, account); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	logger.Info("reconciling account", "account", account.Name)

	if account.DeletionTimestamp != nil {
		delete(r.accounts, account.Status.PublicKey)
		return ctrl.Result{}, nil
	}

	if account.Status.JWT != "" && account.Status.PublicKey != "" {
		r.accounts[account.Status.PublicKey] = account.Status.JWT

		if r.nc != nil {
			go func() {
				if err := r.nc.Publish("$SYS.REQ.CLAIMS.UPDATE", []byte(account.Status.JWT)); err != nil {
					logger.Info("failed to publish claims update", "account", account.Name, "err", err)
				}
			}()
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NatsAccountServer) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&natsv1alpha1.NatsAccount{}).
		Complete(r)
}
