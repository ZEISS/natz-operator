package v1alpha1

import (
	"github.com/nats-io/nkeys"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PrivateKeyPhase string

const (
	PrivateKeyPhaseNone         PrivateKeyPhase = ""
	PrivateKeyPhasePending      PrivateKeyPhase = "Pending"
	PrivateKeyPhaseCreating     PrivateKeyPhase = "Creating"
	PrivateKeyPhaseSynchronized PrivateKeyPhase = "Synchronized"
	PrivateKeyPhaseFailed       PrivateKeyPhase = "Failed"
)

// PrivateKeyType is a type that represents the type of the NKey.
//
// +enum
// +kubebuilder:validation:Enum={Operator,Account,User}
type PrivateKeyType string

// NatsPrivateKeyReference is a reference to a private key
type NatsPrivateKeyReference struct {
	// Name is the name of the private key
	Name string `json:"name"`
}

// NatsPrivateKeySpec defines the desired state of private key
type NatsPrivateKeySpec struct {
	// Type is the type of the NKey.
	Type PrivateKeyType `json:"type"`
}

// NatsPrivateKeyStatus defines the observed state of private key
type NatsPrivateKeyStatus struct {
	// Conditions is an array of conditions that the private key is currently in.
	Conditions []metav1.Condition `json:"conditions,omitempty" optional:"true"`
	// Phase is the current phase of the private key.
	//
	// +kubebuilder:validation:Enum={None,Pending,Creating,Synchronized,Failed}
	Phase PrivateKeyPhase `json:"phase"`
	// ControlPaused is a flag that indicates if the operator is paused.
	ControlPaused bool `json:"controlPaused,omitempty" optional:"true"`
	// LastUpdate is the timestamp of the last update.
	LastUpdate metav1.Time `json:"lastUpdate,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NatsPrivateKey is the Schema for the private key.
type NatsPrivateKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NatsPrivateKeySpec   `json:"spec,omitempty"`
	Status NatsPrivateKeyStatus `json:"status,omitempty"`
}

// Keys returns the keys of the NKey.
func (pk *NatsPrivateKey) Keys() (nkeys.KeyPair, error) {
	var keys nkeys.KeyPair
	var err error

	switch pk.Spec.Type {
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

// NatsPrivateKeyList contains a list of private keys.
type NatsPrivateKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NatsPrivateKey `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NatsPrivateKey{}, &NatsPrivateKeyList{})
}
