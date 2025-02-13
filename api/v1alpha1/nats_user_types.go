package v1alpha1

import (
	"github.com/nats-io/jwt/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// SecretUserJWTKey is the key for the JWT in the secret
	SecretUserJWTKey = "user.jwt"
	// SecretUserCredsKey is the key for the credentials in the secret
	SecretUserCredsKey = "user.creds"
)

type UserPhase string

const (
	UserPhaseNone         UserPhase = ""
	UserPhasePending      UserPhase = "Pending"
	UserPhaseCreating     UserPhase = "Creating"
	UserPhaseSynchronized UserPhase = "Synchronized"
	UserPhaseFailed       UserPhase = "Failed"
)

type Permissions struct {
	Pub  Permission              `json:"pub,omitempty"`
	Sub  Permission              `json:"sub,omitempty"`
	Resp *jwt.ResponsePermission `json:"resp,omitempty"`
}

func (p *Permissions) toNats() jwt.Permissions {
	return jwt.Permissions{
		Pub:  p.Pub.toNats(),
		Sub:  p.Sub.toNats(),
		Resp: p.Resp,
	}
}

type Permission struct {
	Allow jwt.StringList `json:"allow,omitempty"`
	Deny  jwt.StringList `json:"deny,omitempty"`
}

func (p *Permission) toNats() jwt.Permission {
	return jwt.Permission{
		Allow: p.Allow,
		Deny:  p.Deny,
	}
}

// NatsUserSpec defines the desired state of NatsUser
type NatsUserSpec struct {
	// PrivateKey is a reference to a secret that contains the private key
	PrivateKey NatsKeyReference `json:"privateKey,omitempty"`
	// SignerKeyRef is a reference to a secret that contains the account signing key
	SignerKeyRef NatsKeyReference `json:"signerKeyRef"`
	// AccountRef is a reference to the account
	AccountRef NatsReference `json:"accountRef"`
	// Permissions define the permissions for the user
	Permissions Permissions `json:"permissions,omitempty"`
	// Limits define the limits for the user
	Limits Limits `json:"limits,omitempty"`
	// BearerToken is a flag that indicates if the user should be created with a bearer token
	BearerToken bool `json:"bearer_token,omitempty"`
	// AllowedConnectionTypes is a list of allowed connection types
	AllowedConnectionTypes jwt.StringList `json:"allowed_connection_types,omitempty"`
}

type UserLimits struct {
	Src    jwt.CIDRList    `json:"src,omitempty"`
	Times  []jwt.TimeRange `json:"times,omitempty"`
	Locale string          `json:"times_location,omitempty"`
}

func (u *UserLimits) toNats() jwt.UserLimits {
	return jwt.UserLimits{
		Src:    u.Src,
		Times:  u.Times,
		Locale: u.Locale,
	}
}

type Limits struct {
	UserLimits     `json:",inline"`
	jwt.NatsLimits `json:",inline"`
}

func (l *Limits) toNats() jwt.Limits {
	return jwt.Limits{
		UserLimits: l.UserLimits.toNats(),
		NatsLimits: l.NatsLimits,
	}
}

func (s *NatsUserSpec) ToNatsJWT() jwt.User {
	return jwt.User{
		UserPermissionLimits: jwt.UserPermissionLimits{
			Permissions:            s.Permissions.toNats(),
			Limits:                 s.Limits.toNats(),
			BearerToken:            s.BearerToken,
			AllowedConnectionTypes: s.AllowedConnectionTypes,
		},
	}
}

// NatsUserStatus defines the observed state of NatsUser
type NatsUserStatus struct {
	// PublicKey is the public key for the user
	PublicKey string `json:"publicKey,omitempty"`
	// JWT is the JWT for the user
	JWT string `json:"jwt,omitempty"`
	// Conditions is an array of conditions that the operator is currently in.
	Conditions []metav1.Condition `json:"conditions,omitempty" optional:"true"`
	// Phase is the current phase of the operator.
	//
	// +kubebuilder:validation:Enum={None,Pending,Creating,Synchronized,Failed}
	Phase UserPhase `json:"phase"`
	// ControlPaused is a flag that indicates if the operator is paused.
	ControlPaused bool `json:"controlPaused,omitempty" optional:"true"`
	// LastUpdate is the timestamp of the last update.
	LastUpdate metav1.Time `json:"lastUpdate,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +genreconciler
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// NatsUser is the Schema for the natsusers API
type NatsUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NatsUserSpec   `json:"spec,omitempty"`
	Status NatsUserStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NatsUserList contains a list of NatsUser
type NatsUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NatsUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NatsUser{}, &NatsUserList{})
}
