package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Phase is a type that represents the current phase of the operator.
//
// +enum
// +kubebuilder:validation:Enum={None,Pending,Creating,Synchronized,Failed}
type OperatorPhase string

const (
	OperatorPhaseNone         OperatorPhase = ""
	OperatorPhasePending      OperatorPhase = "Pending"
	OperatorPhaseCreating     OperatorPhase = "Creating"
	OperatorPhaseSynchronized OperatorPhase = "Synchronized"
	OperatorPhaseFailed       OperatorPhase = "Failed"
)

// OperatorReference is a reference to an operator.
type NatsOperatorReference struct {
	// Name is the name of the operator.
	Name string `json:"name"`
	// Namespace is the namespace of the operator.
	Namespace string `json:"namespace,omitempty"`
}
type NatsOperatorSpec struct {
	// PrivateKey is a reference to a secret that contains the private key
	PrivateKey NatsKeyReference `json:"privateKey,omitempty"`
	// EnableSystemAccount is a flag that indicates if the system account should be created.
	EnableSystemAccount bool `json:"enableSystemAccount,omitempty"`
	// SigningKeys is a list of references to secrets that contain the signing keys
	SigningKeys []NatsKeyReference `json:"signingKeys,omitempty"`
	// PreventDeletion is a flag that indicates if the  should be locked to prevent deletion.
	// +kubebuilder:default=false
	PreventDeletion bool `json:"prevent_deletion,omitempty"`
	// Paused is a flag that indicates if the  is paused.
	// +kubebuilder:default=false
	Paused bool `json:"paused,omitempty"`
}

type NatsOperatorStatus struct {
	// JWT is the JWT that the operator is currently using.
	JWT string `json:"jwt"`
	// PublicKey is the public key that the operator is currently using.
	PublicKey string `json:"publicKey"`
	// Conditions is an array of conditions that the operator is currently in.
	Conditions []metav1.Condition `json:"conditions,omitempty" optional:"true"`
	// Phase is the current phase of the operator.
	//
	// +kubebuilder:validation:Enum={None,Pending,Creating,Synchronized,Failed}
	Phase OperatorPhase `json:"phase"`
	// ControlPaused is a flag that indicates if the operator is paused.
	ControlPaused bool `json:"controlPaused,omitempty" optional:"true"`
	// LastUpdate is the timestamp of the last update.
	LastUpdate metav1.Time `json:"lastUpdate,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type NatsOperator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NatsOperatorSpec   `json:"spec,omitempty"`
	Status NatsOperatorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NatsOperatorList contains a list of NatsOperator
type NatsOperatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NatsOperator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NatsOperator{}, &NatsOperatorList{})
}
