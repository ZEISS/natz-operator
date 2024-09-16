package controllers

import (
	"context"
	"fmt"
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/strings/slices"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	natsv1alpha1 "github.com/zeiss/natz-operator/api/v1alpha1"
)

const ACCOUNT_TEMPLATE = `-----BEGIN NATS USER JWT-----
%s
------END NATS USER JWT------

-----BEGIN USER NKEY SEED-----
%s
------END USER NKEY SEED------
`

// NatsUserReconciler reconciles a NatsUser object
type NatsUserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// NewNatsUserReconciler ...
func NewNatsUserReconciler(mgr ctrl.Manager) *NatsUserReconciler {
	return &NatsUserReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
}

//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsusers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsusers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsusers/finalizers,verbs=update

// Reconcile ...
// nolint:gocyclo
func (r *NatsUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	user := &natsv1alpha1.NatsUser{}
	if err := r.Get(ctx, req.NamespacedName, user); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if user.DeletionTimestamp != nil {
		logger.Info("Processing deletion of user")

		if controllerutil.RemoveFinalizer(user, NATZ_OPERATOR_FINALIZER) {
			if err := r.Update(ctx, user); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	if controllerutil.AddFinalizer(user, NATZ_OPERATOR_FINALIZER) {
		if err := r.Update(ctx, user); err != nil {
			return ctrl.Result{}, err
		}
	}

	issuingAccount := &natsv1alpha1.NatsAccount{}
	signerSecret := &corev1.Secret{}
	for {
		if err := r.Get(ctx, client.ObjectKey{
			Namespace: user.Spec.AccountRef.Namespace,
			Name:      user.Spec.AccountRef.Name,
		}, issuingAccount); err != nil {
			return ctrl.Result{}, err
		}

		if issuingAccount.Status.AccountSecretName == "" {
			logger.Info("waiting for issuing account secret to appear")
			<-time.After(5 * time.Second)
			continue
		}

		if !slices.Contains(issuingAccount.Spec.AllowUserNamespaces, req.Namespace) {
			return ctrl.Result{}, nil
		}

		if err := r.Get(ctx, client.ObjectKey{
			Namespace: issuingAccount.Namespace,
			Name:      issuingAccount.Status.AccountSecretName,
		}, signerSecret); err != nil {
			return ctrl.Result{}, err
		}

		break
	}

	_, err := r.reconcileSecret(ctx, req, user, signerSecret)
	return ctrl.Result{}, err
}

// nolint:gocyclo
func (r *NatsUserReconciler) reconcileSecret(ctx context.Context, req ctrl.Request, user *natsv1alpha1.NatsUser, signerSecret *corev1.Secret) (*corev1.Secret, error) {
	logger := log.FromContext(ctx)
	keySecret := &corev1.Secret{}
	hasSecret := true
	hasChanges := false

	if err := r.Get(ctx, req.NamespacedName, keySecret); errors.IsNotFound(err) {
		keySecret.Namespace = req.Namespace
		keySecret.Name = req.Name
		keySecret.Type = "natz.zeiss.com/nats-user"
		hasSecret = false
		if err := controllerutil.SetOwnerReference(user, keySecret, r.Scheme); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	logger.Info("reconciling user keys")
	hasChanges, err := r.reconcileKey(ctx, keySecret, user, signerSecret.Data[OPERATOR_SEED_KEY])
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
		// Update operator status if we encountered changes
		user.Status.UserSecretName = keySecret.Name
		user.Status.PublicKey = string(keySecret.Data[OPERATOR_PUBLIC_KEY])
		user.Status.JWT = string(keySecret.Data[OPERATOR_JWT])
		if err := r.Status().Update(ctx, user); err != nil {
			return nil, err
		}
	}
	return keySecret, nil
}

// nolint:gocyclo
func (r *NatsUserReconciler) reconcileKey(ctx context.Context, secret *corev1.Secret, account *natsv1alpha1.NatsUser, signer []byte) (bool, error) {
	logger := log.FromContext(ctx)
	keys, needsKeyUpdate, err := extractOrCreateKeys(secret, nkeys.CreateUser)
	if err != nil {
		return false, err
	}

	seed, _ := keys.Seed()
	public, _ := keys.PublicKey()

	token := jwt.NewUserClaims(public)
	token.User = account.Spec.ToNatsJWT()
	needsClaimsUpdate := secret.Data == nil
	signerKp, err := nkeys.FromSeed(signer)
	if err != nil {
		return false, fmt.Errorf("failed decoding seed: %w, signer: %v", err, signer)
	}

	if secret.Data != nil {
		oldToken, err := jwt.DecodeUserClaims(string(secret.Data[OPERATOR_JWT]))
		if err == nil {
			needsClaimsUpdate = needsClaimsUpdate || !reflect.DeepEqual(token.User, oldToken.User)
			// Check if the signing keys changed
			needsClaimsUpdate = needsClaimsUpdate || oldToken.Issuer != token.Issuer
		} else {
			// Claims could not be decoded, need update.
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
			logger.Info("token encode error", "pubkey", token.Subject, "public", public)
			return false, err
		}
		secret.Data[OPERATOR_JWT] = []byte(jwt)
		secret.Data[OPERATOR_CREDS] = []byte(fmt.Sprintf(ACCOUNT_TEMPLATE, jwt, seed))
	}
	return needsKeyUpdate || needsClaimsUpdate, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NatsUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&natsv1alpha1.NatsUser{}).
		Complete(r)
}
