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

type NatsOperatorSpec struct {
	// PrivateKey is a reference to a secret that contains the private key
	PrivateKey NatsPrivateKeyReference `json:"private_key,omitempty"`
	// SigningKeys is a list of references to secrets that contain the signing keys
	SigningKeys []NatsSigningKeyReference `json:"signing_keys,omitempty"`
}

type NatsOperatorStatus struct {
	JWT        string `json:"jwt"`
	PublicKey  string `json:"publicKey"`
	SecretName string `json:"secretName"`

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
