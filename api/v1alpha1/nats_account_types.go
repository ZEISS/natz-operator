package v1alpha1

import (
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/samber/lo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AccountPhase string

const (
	AccountPhaseNone         AccountPhase = ""
	AccountPhasePending      AccountPhase = "Pending"
	AccountPhaseCreating     AccountPhase = "Creating"
	AccountPhaseSynchronized AccountPhase = "Synchronized"
	AccountPhaseFailed       AccountPhase = "Failed"
)

// NatsAccountReference is a reference to a NatsAccount
type NatsAccountReference struct {
	// Name is the name of the account.
	Name string `json:"name"`
	// Namespace is the namespace of the account.
	Namespace string `json:"namespace,omitempty"`
}

// ExportType defines the type of import/export.
type ExportType int

const (
	// Unknown is used if we don't know the type
	Unknown ExportType = iota
	// Stream defines the type field value for a stream "stream"
	Stream
	// Service defines the type field value for a service "service"
	Service
)

// Export ...
type Export struct {
	Name                 string              `json:"name,omitempty"`
	Subject              jwt.Subject         `json:"subject,omitempty"`
	Type                 ExportType          `json:"type,omitempty"`
	TokenReq             bool                `json:"token_req,omitempty"`
	Revocations          jwt.RevocationList  `json:"revocations,omitempty"`
	ResponseType         jwt.ResponseType    `json:"response_type,omitempty"`
	ResponseThreshold    time.Duration       `json:"response_threshold,omitempty"`
	Latency              *jwt.ServiceLatency `json:"service_latency,omitempty"`
	AccountTokenPosition uint                `json:"account_token_position,omitempty"`
	Advertise            bool                `json:"advertise,omitempty"`
	jwt.Info             `json:",inline"`
}

// OperatorLimits are used to limit access by an account
type OperatorLimits struct {
	jwt.NatsLimits            `json:",inline"`
	jwt.AccountLimits         `json:",inline"`
	jwt.JetStreamLimits       `json:",inline"`
	jwt.JetStreamTieredLimits `json:"tiered_limits,omitempty"`
}

// NatsAccountSpec defines the desired state of NatsAccount
type NatsAccountSpec struct {
	// SignerKeyRef is the reference to the secret that contains the signing key
	SignerKeyRef NatsKeyReference `json:"signerKeyRef,omitempty"`
	// PrivateKey is a reference to a secret that contains the private key
	PrivateKey NatsKeyReference `json:"privateKey,omitempty"`
	// SigningKeys is a list of references to secrets that contain the signing keys
	SigningKeys []NatsKeyReference `json:"signingKeys,omitempty"`
	// OperatorSigningKey is the reference to the operator signing key
	OperatorSigningKey NatsKeyReference `json:"operatorSigningKey,omitempty"`
	// Namespaces that are allowed for user creation.
	// If a NatsUser is referencing this account outside of these namespaces, the operator will create an event for it saying that it's not allowed.
	AllowUserNamespaces []string `json:"allowedUserNamespaces,omitempty"`

	// These fields are directly mappejwtd into the NATS JWT claim
	Imports     []*jwt.Import      `json:"imports,omitempty"`
	Exports     []Export           `json:"exports,omitempty"`
	Limits      OperatorLimits     `json:"limits,omitempty"`
	Revocations jwt.RevocationList `json:"revocations,omitempty"`
}

func (s *NatsAccountSpec) ToJWTAccount() jwt.Account {
	exports := lo.Map(s.Exports, func(e Export, _ int) *jwt.Export {
		return &jwt.Export{
			Name:                 e.Name,
			Subject:              e.Subject,
			Type:                 jwt.ExportType(e.Type),
			TokenReq:             e.TokenReq,
			Revocations:          e.Revocations,
			ResponseType:         e.ResponseType,
			ResponseThreshold:    e.ResponseThreshold,
			Latency:              e.Latency,
			AccountTokenPosition: e.AccountTokenPosition,
			Advertise:            e.Advertise,
			Info:                 e.Info,
		}
	})

	return jwt.Account{
		Imports: jwt.Imports(s.Imports),
		Exports: jwt.Exports(exports),
		Limits: jwt.OperatorLimits{
			NatsLimits:            s.Limits.NatsLimits,
			AccountLimits:         s.Limits.AccountLimits,
			JetStreamLimits:       s.Limits.JetStreamLimits,
			JetStreamTieredLimits: s.Limits.JetStreamTieredLimits,
		},
		SigningKeys: jwt.SigningKeys{},
		Revocations: s.Revocations,
	}
}

// NatsAccountStatus defines the observed state of NatsAccount
type NatsAccountStatus struct {
	// PublicKey is the public key that the account is currently using.
	PublicKey string `json:"publicKey,omitempty"`
	// JWT is the JWT that the account is currently using.
	JWT string `json:"jwt,omitempty"`
	// Conditions is an array of conditions that the operator is currently in.
	Conditions []metav1.Condition `json:"conditions,omitempty" optional:"true"`
	// Phase is the current phase of the operator.
	//
	// +kubebuilder:validation:Enum={None,Pending,Creating,Synchronized,Failed}
	Phase AccountPhase `json:"phase"`
	// ControlPaused is a flag that indicates if the operator is paused.
	ControlPaused bool `json:"controlPaused,omitempty" optional:"true"`
	// LastUpdate is the timestamp of the last update.
	LastUpdate metav1.Time `json:"lastUpdate,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NatsAccount is the Schema for the natsaccounts API
type NatsAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NatsAccountSpec   `json:"spec,omitempty"`
	Status NatsAccountStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NatsAccountList contains a list of NatsAccount
type NatsAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NatsAccount `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NatsAccount{}, &NatsAccountList{})
}
