package controllers

import (
	"context"
	"fmt"
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	natsv1alpha1 "github.com/zeiss/natz-operator/api/v1alpha1"
)

// NatsAccountReconciler ...
type NatsAccountReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// NewNatsAccountReconciler ...
func NewNatsAccountReconciler(mgr ctrl.Manager) *NatsAccountReconciler {
	return &NatsAccountReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
}

//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsaccounts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsaccounts/finalizers,verbs=update

// Reconcile ...
// nolint:gocyclo
func (r *NatsAccountReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	account := &natsv1alpha1.NatsAccount{}
	if err := r.Get(ctx, req.NamespacedName, account); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if account.DeletionTimestamp != nil {
		logger.Info("Processing deletion of account")

		if controllerutil.RemoveFinalizer(account, NATZ_OPERATOR_FINALIZER) {
			if err := r.Update(ctx, account); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	issuer := &natsv1alpha1.NatsOperator{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      account.Spec.OperatorRef.Name,
	}, issuer); errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	if err := controllerutil.SetOwnerReference(issuer, account, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	if controllerutil.AddFinalizer(account, NATZ_OPERATOR_FINALIZER) {
		if err := r.Update(ctx, account); err != nil {
			return ctrl.Result{}, err
		}
	}

	signerSecret := &corev1.Secret{}
	for {
		if issuer.Status.OperatorSecretName == "" {
			logger.Info("waiting for issuing account secret to appear")

			<-time.After(5 * time.Second)

			continue
		}

		if err := r.Get(ctx, client.ObjectKey{
			Namespace: issuer.Namespace,
			Name:      issuer.Status.OperatorSecretName,
		}, signerSecret); err != nil {
			return ctrl.Result{}, err
		}

		logger.Info("issuing account secret found")

		break
	}

	_, err := r.reconcileSecret(ctx, req, account, signerSecret)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// nolint:gocyclo
func (r *NatsAccountReconciler) reconcileSecret(ctx context.Context, req ctrl.Request, account *natsv1alpha1.NatsAccount, signerSecret *corev1.Secret) (*corev1.Secret, error) {
	// Try reconcile the secret containing the seed key for the operator
	logger := log.FromContext(ctx)
	keySecret := &corev1.Secret{}
	hasSecret := true
	hasChanges := false
	if err := r.Get(ctx, req.NamespacedName, keySecret); errors.IsNotFound(err) {
		keySecret.Namespace = req.Namespace
		keySecret.Name = req.Name
		keySecret.Type = "natz.zeiss.com/nats-account"
		hasSecret = false

		if err := controllerutil.SetOwnerReference(account, keySecret, r.Scheme); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	logger.Info("reconciling account keys")

	hasChanges, err := r.reconcileKey(ctx, keySecret, account, signerSecret.Data[OPERATOR_SEED_KEY])
	if err != nil {
		return nil, err
	}

	if !hasSecret {
		if err := r.Create(ctx, keySecret); err != nil {
			return nil, err
		}
	} else if hasChanges {
		if err := r.Update(ctx, keySecret); err != nil {
			return nil, err
		}
	}

	if !hasSecret || hasChanges {
		account.Status.AccountSecretName = keySecret.Name
		account.Status.PublicKey = string(keySecret.Data[OPERATOR_PUBLIC_KEY])
		account.Status.JWT = string(keySecret.Data[OPERATOR_JWT])

		if err := r.Status().Update(ctx, account); err != nil {
			return nil, err
		}
	}
	return keySecret, nil
}

// nolint:gocyclo
func (r *NatsAccountReconciler) reconcileKey(ctx context.Context, secret *corev1.Secret, account *natsv1alpha1.NatsAccount, signer []byte) (bool, error) {
	logger := log.FromContext(ctx)
	keys, needsKeyUpdate, err := extractOrCreateKeys(secret, nkeys.CreateAccount)
	if err != nil {
		return false, err
	}

	seed, _ := keys.Seed()
	public, _ := keys.PublicKey()

	token := jwt.NewAccountClaims(public)
	token.Account = account.Spec.ToJWTAccount()
	needsClaimsUpdate := secret.Data == nil
	signerKp, err := nkeys.FromSeed(signer)
	if err != nil {
		return false, fmt.Errorf("failed decoding seed: %w, signer: %v", err, signer)
	}

	if secret.Data != nil {
		oldToken, err := jwt.DecodeAccountClaims(string(secret.Data[OPERATOR_JWT]))
		if err == nil {
			needsClaimsUpdate = needsClaimsUpdate || !reflect.DeepEqual(token.Account, oldToken.Account)
			needsClaimsUpdate = needsClaimsUpdate || oldToken.Issuer != token.Issuer
		} else {
			needsClaimsUpdate = true
		}
	}

	logger.Info("updating secret if needed", "needsUpdate", needsClaimsUpdate)

	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}

	if needsKeyUpdate {
		secret.Data[OPERATOR_SEED_KEY] = seed
		secret.Data[OPERATOR_PUBLIC_KEY] = []byte(public)
	}

	if needsKeyUpdate || needsClaimsUpdate {
		jwt, err := token.Encode(signerKp)
		if err != nil {
			return false, err
		}
		secret.Data[OPERATOR_JWT] = []byte(jwt)
	}

	return needsKeyUpdate || needsClaimsUpdate, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NatsAccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&natsv1alpha1.NatsAccount{}).
		Complete(r)
}
