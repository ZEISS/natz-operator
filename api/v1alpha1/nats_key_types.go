package v1alpha1

import (
	"github.com/nats-io/nkeys"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NatsKeyReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

type KeyPhase string

const (
	KeyPhaseNone         KeyPhase = ""
	KeyPhasePending      KeyPhase = "Pending"
	KeyPhaseCreating     KeyPhase = "Creating"
	KeyPhaseSynchronized KeyPhase = "Synchronized"
	KeyPhaseFailed       KeyPhase = "Failed"
)

// KeyType is a type that represents the type of the N.
//
// +enum
// +kubebuilder:validation:Enum={Operator,Account,User}
type KeyType string

// NatsReference is a reference to a .
type NatsReference struct {
	// Name is the name of the
	Name string `json:"name"`
	// Namespace is the namespace of the private
	Namespace string `json:"namespace,omitempty"`
}

// NatsPrivateSpec defines the desired state of private
type NatsPrivateSpec struct {
	// Type is the type of the N.
	Type KeyType `json:"type"`
	// PreventDeletion is a flag that indicates if the  should be locked to prevent deletion.
	PreventDeletion bool `json:"prevent_deletion,omitempty"`
}

// NatsPrivateStatus defines the observed state of private
type NatsPrivateStatus struct {
	// Conditions is an array of conditions that the private  is currently in.
	Conditions []metav1.Condition `json:"conditions,omitempty" optional:"true"`
	// Phase is the current phase of the private .
	//
	// +kubebuilder:validation:Enum={None,Pending,Creating,Synchronized,Failed}
	Phase KeyPhase `json:"phase"`
	// ControlPaused is a flag that indicates if the operator is paused.
	ControlPaused bool `json:"controlPaused,omitempty" optional:"true"`
	// LastUpdate is the timestamp of the last update.
	LastUpdate metav1.Time `json:"lastUpdate,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NatsKey is the Schema for the key.
type NatsKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NatsPrivateSpec   `json:"spec,omitempty"`
	Status NatsPrivateStatus `json:"status,omitempty"`
}

// Keys returns a pair of keys based on the type of the N.
func (pk *NatsKey) Keys() (nkeys.KeyPair, error) {
	var s nkeys.KeyPair
	var err error

	switch pk.Spec.Type {
	case "Operator":
		s, err = nkeys.CreateOperator()
	case "Account":
		s, err = nkeys.CreateAccount()
	case "User":
		s, err = nkeys.CreateUser()
	}

	return s, err
}

//+kubebuilder:object:root=true

// NatsKeyList contains a list of key.
type NatsKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NatsKey `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NatsKey{}, &NatsKeyList{})
}
