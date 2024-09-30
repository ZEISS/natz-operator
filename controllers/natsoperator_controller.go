package controllers

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	natsv1alpha1 "github.com/zeiss/natz-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// NatsOperatorReconciler ...
type NatsOperatorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// NewNatsOperatorReconciler ...
func NewNatsOperatorReconciler(mgr ctrl.Manager) *NatsOperatorReconciler {
	return &NatsOperatorReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
}

//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsoperators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsoperators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=natz.zeiss.com,resources=natsoperators/finalizers,verbs=update
//+kubebuilder:rbac:groups=,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile ...
// nolint:gocyclo
func (r *NatsOperatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	operator := &natsv1alpha1.NatsOperator{}
	if err := r.Get(ctx, req.NamespacedName, operator); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if operator.DeletionTimestamp != nil {
		logger.Info("Processing deletion of operator")

		if controllerutil.RemoveFinalizer(operator, NATZ_OPERATOR_FINALIZER) {
			if err := r.Update(ctx, operator); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	if controllerutil.AddFinalizer(operator, NATZ_OPERATOR_FINALIZER) {
		if err := r.Update(ctx, operator); err != nil {
			return ctrl.Result{}, err
		}
	}

	needsRewriteConfig, err := r.reconcileSecret(ctx, req, operator)
	if err != nil {
		return ctrl.Result{}, err
	}

	systemAccount := &natsv1alpha1.NatsAccount{}
	systemAccountName := client.ObjectKey{
		Namespace: req.Namespace,
		Name:      fmt.Sprintf("%v-system", req.Name),
	}

	for {
		// nolint:gocritic
		if err := r.Get(ctx, systemAccountName, systemAccount); errors.IsNotFound(err) {
			logger.Info("creating system account")
			systemAccount.Name = systemAccountName.Name
			systemAccount.Namespace = systemAccountName.Namespace
			if err := controllerutil.SetOwnerReference(operator, systemAccount, r.Scheme); err != nil {
				return ctrl.Result{}, err
			}

			systemAccount.Spec = natsv1alpha1.NatsAccountSpec{
				AllowUserNamespaces: []string{req.Namespace},
				OperatorRef: corev1.ObjectReference{
					Namespace: req.Namespace,
					Name:      req.Name,
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

			if err := r.Create(ctx, systemAccount); err != nil {
				for _, e := range systemAccount.Spec.Exports {
					logger.Info("export", "name", e.Name, "subject", e.Subject, "type", e.Type)
				}
				return ctrl.Result{}, err
			}

			// After the user has been created, we need to requeue this operator nkey
			// because we need to enqueue the account and the user in order to create the server config
			return ctrl.Result{
				RequeueAfter: 5 * time.Second,
			}, nil
		} else if err != nil {
			return ctrl.Result{}, err
		} else if systemAccount.Status.JWT == "" {
			// Object has been found, but JWT hasn't been issued, wait until it has been issued
			logger.Info("waiting for system account to become ready")
			<-time.After(5 * time.Second)
		} else {
			break
		}
	}

	systemUser := &natsv1alpha1.NatsUser{}
	systemUserName := client.ObjectKey{
		Namespace: req.Namespace,
		Name:      fmt.Sprintf("%v-jwt", req.Name),
	}

	for {
		// nolint:gocritic
		if err := r.Get(ctx, systemUserName, systemUser); errors.IsNotFound(err) {
			logger.Info("creating jwt system user")
			systemUser.Name = systemUserName.Name
			systemUser.Namespace = systemUserName.Namespace
			if err := controllerutil.SetOwnerReference(operator, systemUser, r.Scheme); err != nil {
				return ctrl.Result{}, err
			}
			// Allow this user to publish and subscribe, i.e. interact with the server for JWT permissions
			systemUser.Spec = natsv1alpha1.NatsUserSpec{
				AccountRef: corev1.ObjectReference{
					Namespace: systemAccount.Namespace,
					Name:      systemAccount.Name,
				},
				Limits: natsv1alpha1.Limits{
					NatsLimits: jwt.NatsLimits{
						Subs:    -1,
						Payload: -1,
						Data:    -1,
					},
				},
			}
			if err := r.Create(ctx, systemUser); err != nil {
				return ctrl.Result{}, err
			}

			// After the user has been created, we need to requeue this operator nkey
			// because we need to enqueue the account and the user in order to create the server config
			return ctrl.Result{
				RequeueAfter: 5 * time.Second,
			}, nil
		} else if err != nil {
			return ctrl.Result{}, err
		} else if systemUser.Status.JWT == "" {
			// Object has been found but jwt not issued yet
			logger.Info("waiting for system user to become ready")
			<-time.After(5 * time.Second)
		} else {
			break
		}
	}

	return ctrl.Result{}, r.reconcileServerConfigSnipped(ctx, req, operator, systemAccount, needsRewriteConfig)
}

func (r *NatsOperatorReconciler) reconcileServerConfigSnipped(ctx context.Context, req ctrl.Request, operator *natsv1alpha1.NatsOperator, sysacc *natsv1alpha1.NatsAccount, needsRefresh bool) error {
	logger := log.FromContext(ctx)
	// Finally, reconcile server configuration snippet
	serverConfig := &corev1.Secret{}
	hasSecret := true
	serverConfigName := client.ObjectKey{
		Namespace: req.Namespace,
		Name:      fmt.Sprintf("%v-server-config", req.Name),
	}
	if err := r.Get(ctx, serverConfigName, serverConfig); errors.IsNotFound(err) {
		logger.Info("creating server config")
		serverConfig.Namespace = req.Namespace
		serverConfig.Name = serverConfigName.Name
		serverConfig.Type = "natz.zeiss.com/nats-configuration"
		hasSecret = false
	} else if err != nil {
		return err
	}

	text := fmt.Sprintf(AUTH_CONFIG_TEMPLATE, operator.Status.JWT, sysacc.Status.PublicKey, sysacc.Status.PublicKey, sysacc.Status.JWT)
	if !needsRefresh && serverConfig.Data != nil {
		needsRefresh = needsRefresh || text != string(serverConfig.Data[OPERATOR_CONFIG_FILE])
	}

	if serverConfig.Data == nil || needsRefresh {
		serverConfig.Data = map[string][]byte{
			OPERATOR_CONFIG_FILE: []byte(text),
		}
	}

	if err := controllerutil.SetOwnerReference(operator, serverConfig, r.Scheme); err != nil {
		return err
	}

	if !hasSecret {
		return r.Create(ctx, serverConfig)
	}

	return r.Update(ctx, serverConfig)
}

// nolint:gocyclo
func (r *NatsOperatorReconciler) reconcileSecret(ctx context.Context, req ctrl.Request, operator *natsv1alpha1.NatsOperator) (bool, error) {
	// Try reconcile the secret containing the seed key for the operator
	logger := log.FromContext(ctx)
	operatorKeySecret := &corev1.Secret{}
	hasSecret := true
	hasChanges := false
	if err := r.Get(ctx, req.NamespacedName, operatorKeySecret); errors.IsNotFound(err) {
		operatorKeySecret.Namespace = req.Namespace
		operatorKeySecret.Name = req.Name
		operatorKeySecret.Type = "natz.zeiss.com/nats-operator"
		hasSecret = false

		if err := controllerutil.SetOwnerReference(operator, operatorKeySecret, r.Scheme); err != nil {
			return false, err
		}

	} else if err != nil {
		return false, err
	}

	logger.Info("reconciling operator keys")
	hasChanges, err := r.reconcileKey(ctx, operatorKeySecret, operator)
	if err != nil {
		return false, err
	}

	if !hasSecret {
		if err := r.Create(ctx, operatorKeySecret); err != nil {
			return false, err
		}
	} else if hasChanges {
		if err := r.Update(ctx, operatorKeySecret); err != nil {
			return false, err
		}
	}

	if !hasSecret || hasChanges {
		operator.Status.OperatorSecretName = operatorKeySecret.Name
		operator.Status.PublicKey = string(operatorKeySecret.Data[OPERATOR_PUBLIC_KEY])
		operator.Status.JWT = string(operatorKeySecret.Data[OPERATOR_JWT])
		if err := r.Status().Update(ctx, operator); err != nil {
			return false, err
		}
	}
	return !hasSecret || hasChanges, nil
}

// nolint:gocyclo
func (r *NatsOperatorReconciler) reconcileKey(ctx context.Context, secret *corev1.Secret, operator *natsv1alpha1.NatsOperator) (bool, error) {
	logger := log.FromContext(ctx)
	keys, needsKeyUpdate, err := extractOrCreateKeys(secret, nkeys.CreateOperator)
	if err != nil {
		return false, err
	}

	seed, _ := keys.Seed()
	public, _ := keys.PublicKey()

	token := jwt.NewOperatorClaims(public)
	token.Operator.SigningKeys = operator.Spec.SigningKeys
	needsClaimsUpdate := secret.Data == nil

	if secret.Data != nil {
		oldToken, err := jwt.DecodeOperatorClaims(string(secret.Data[OPERATOR_JWT]))
		if err == nil {
			needsClaimsUpdate = needsClaimsUpdate || !reflect.DeepEqual(token.Operator.SigningKeys, oldToken.Operator.SigningKeys)
		} else {
			needsClaimsUpdate = true
		}
	}

	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}

	logger.Info("updating secret if needed", "claimsChanged", needsClaimsUpdate)

	if needsKeyUpdate {
		secret.Data[OPERATOR_SEED_KEY] = seed
		secret.Data[OPERATOR_PUBLIC_KEY] = []byte(public)
	}

	if needsKeyUpdate || needsClaimsUpdate {
		// Whenerver our keys changed, we also need to force renew the token
		jwt, err := token.Encode(keys)
		if err != nil {
			return false, err
		}
		secret.Data[OPERATOR_JWT] = []byte(jwt)
	}

	return needsKeyUpdate || needsClaimsUpdate, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NatsOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&natsv1alpha1.NatsOperator{}).
		Complete(r)
}

// nolint:gocyclo
func extractOrCreateKeys(secret *corev1.Secret, generator func() (nkeys.KeyPair, error)) (nkeys.KeyPair, bool, error) {
	var keys nkeys.KeyPair
	needsKeyUpdate := true

	if secret.Data != nil {
		parsedKeys, err := nkeys.FromSeed(secret.Data[OPERATOR_SEED_KEY])
		if err == nil {
			keys = parsedKeys
			needsKeyUpdate = false
		}
	}

	if keys == nil {
		// No keys present or failed to extract, create new key pair
		createdKeys, err := generator()
		if err != nil {
			return nil, false, err
		}
		keys = createdKeys
	}

	return keys, needsKeyUpdate, nil
}
