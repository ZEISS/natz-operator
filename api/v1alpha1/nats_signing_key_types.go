package v1alpha1

import (
	"github.com/nats-io/nkeys"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SigningKeyPhase string

const (
	SigningKeyPhaseNone         SigningKeyPhase = ""
	SigningKeyPhasePending      SigningKeyPhase = "Pending"
	SigningKeyPhaseCreating     SigningKeyPhase = "Creating"
	SigningKeyPhaseSynchronized SigningKeyPhase = "Synchronized"
	SigningKeyPhaseFailed       SigningKeyPhase = "Failed"
)

// SigningKeyType is a type that represents the type of the NKey.
//
// +enum
// +kubebuilder:validation:Enum={Operator,Account,User}
type SigningKeyType string

// NatsSigningKeyReference is a reference to a NatsSigningKey
type NatsSigningKeyReference struct {
	// Name is the name of the NatsSigningKey
	Name string `json:"name"`
	// Namespace is the namespace of the NatsSigningKey
	Namespace string `json:"namespace,omitempty"`
}

// NatsSigningKeySpec defines the desired state of SigningKey
type NatsSigningKeySpec struct {
	// Type is the type of the NKey.
	Type SigningKeyType `json:"type"`
	// PreventDeletion is a flag that indicates if the key should be locked to prevent deletion.
	PreventDeletion bool `json:"prevent_deletion,omitempty"`
}

// NatsSigningKeyStatus defines the observed state of SigningKey
type NatsSigningKeyStatus struct {
	// Conditions is an array of conditions that the operator is currently in.
	Conditions []metav1.Condition `json:"conditions,omitempty" optional:"true"`
	// Phase is the current phase of the operator.
	//
	// +kubebuilder:validation:Enum={None,Pending,Creating,Synchronized,Failed}
	Phase SigningKeyPhase `json:"phase"`
	// ControlPaused is a flag that indicates if the operator is paused.
	ControlPaused bool `json:"controlPaused,omitempty" optional:"true"`
	// LastUpdate is the timestamp of the last update.
	LastUpdate metav1.Time `json:"lastUpdate,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NatsSigningKey is the Schema for the natssigningkeys API
type NatsSigningKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NatsSigningKeySpec   `json:"spec,omitempty"`
	Status NatsSigningKeyStatus `json:"status,omitempty"`
}

// Keys returns the keys of the NKey.
func (sk *NatsSigningKey) Keys() (nkeys.KeyPair, error) {
	var keys nkeys.KeyPair
	var err error

	switch sk.Spec.Type {
	case "Operator":
		keys, err = nkeys.CreateOperator()
	case "Account":
		keys, err = nkeys.CreateAccount()
	case "User":
		keys, err = nkeys.CreateUser()
	}

	return keys, err
}

//+kubebuilder:object:root=true

// NatsSigningKeyList contains a list of NatsSigningKey
type NatsSigningKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NatsSigningKey `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NatsSigningKey{}, &NatsSigningKeyList{})
}
