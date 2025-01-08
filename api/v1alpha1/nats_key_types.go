package v1alpha1

import (
	"errors"

	"github.com/nats-io/nkeys"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NatsKeyReference struct {
	// Name is the name of the key as a reference
	Name string `json:"name"`
	// Namespace is the namespace of the key as a reference
	Namespace string `json:"namespace,omitempty"`
}

// NatsKeyPhase is a type that represents the phase of the N.
type NatsKeyPhase string

const (
	KeyPhaseNone         NatsKeyPhase = ""
	KeyPhasePending      NatsKeyPhase = "Pending"
	KeyPhaseCreating     NatsKeyPhase = "Creating"
	KeyPhaseSynchronized NatsKeyPhase = "Synchronized"
	KeyPhaseFailed       NatsKeyPhase = "Failed"
)

// KeyType is a type that represents the type of the N.
//
// +enum
// +kubebuilder:validation:Enum={Operator,Account,User}
type KeyType string

const (
	KeyTypeOperator KeyType = "Operator"
	KeyTypeAccount  KeyType = "Account"
	KeyTypeUser     KeyType = "User"
)

var ErrUnknownKeyType = errors.New("unknown key type")

// NatsReference is a reference to a .
type NatsReference struct {
	// Name is the name of the
	Name string `json:"name"`
	// Namespace is the namespace of the private
	Namespace string `json:"namespace,omitempty"`
}

// NatsKeySpec defines the desired state of a NATS key.
type NatsKeySpec struct {
	// Type is the type of the N.
	Type KeyType `json:"type"`
	// PreventDeletion is a flag that indicates if the  should be locked to prevent deletion.
	PreventDeletion bool `json:"prevent_deletion,omitempty"`
	// Paused is a flag that indicates if the  is paused.
	// +kubebuilder:default=false
	Paused bool `json:"paused,omitempty"`
}

// NatsKeyStatus defines the observed state of a NATS key.
type NatsKeyStatus struct {
	// Conditions is an array of conditions that the private  is currently in.
	Conditions []metav1.Condition `json:"conditions,omitempty" optional:"true"`
	// Phase is the current phase of the private .
	//
	// +kubebuilder:validation:Enum={None,Pending,Creating,Synchronized,Failed}
	Phase NatsKeyPhase `json:"phase"`
	// ControlPaused is a flag that indicates if the operator is paused.
	ControlPaused bool `json:"controlPaused,omitempty" optional:"true"`
	// LastUpdate is the timestamp of the last update.
	LastUpdate metav1.Time `json:"lastUpdate,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NatsKey is the Schema for a NATS key.
type NatsKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NatsKeySpec   `json:"spec,omitempty"`
	Status NatsKeyStatus `json:"status,omitempty"`
}

// Keys returns a pair of keys based on the type of the N.
func (pk *NatsKey) Keys() (nkeys.KeyPair, error) {
	var s nkeys.KeyPair
	var err error

	switch pk.Spec.Type {
	case KeyTypeOperator:
		s, err = nkeys.CreateOperator()
	case KeyTypeAccount:
		s, err = nkeys.CreateAccount()
	case KeyTypeUser:
		s, err = nkeys.CreateUser()
	default:
		err = ErrUnknownKeyType
	}

	return s, err
}

//+kubebuilder:object:root=true

// NatsKeyList contains a list of NATS keys.
type NatsKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NatsKey `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NatsKey{}, &NatsKeyList{})
}
