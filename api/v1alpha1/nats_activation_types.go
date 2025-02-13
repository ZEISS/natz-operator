package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ActivationPhase string

const (
	ActivationPhaseNone     ActivationPhase = ""
	ActivationPhaseCreating ActivationPhase = "Creating"
	ActivationPhaseActive   ActivationPhase = "Active"
	ActivationSynchronized  ActivationPhase = "Synchronized"
	ActivationPhaseFailed   ActivationPhase = "Failed"
)

// NatsActivationRef is a reference to a NatsActivation
type NatsActivationReference struct {
	// Name is the name of the Activation
	Name string `json:"name"`
	// Namespace is the namespace of the Activation
	Namespace string `json:"namespace"`
}

// NatsActivationSpec defines the desired state of NatsActivation
type NatsActivationSpec struct {
	// AccountRef is a reference to the account that the activation is for.
	AccountRef NatsAccountReference `json:"accountRef"`
	// SignerKeyRef is a reference to a secret that contains the account signing key
	SignerKeyRef NatsKeyReference `json:"signerKeyRef"`
	// TargetAccountRef is a reference to the account that the activation is for.
	TargetAccountRef NatsAccountReference `json:"targetAccountRef"`
	// Expiry is the expiry time of the activation.
	Expiry metav1.Time `json:"expiry,omitempty"`
	// Start is the start time of the activation.
	Start metav1.Time `json:"start,omitempty"`
	// Subject is the subject that the activation is for.
	Subject string `json:"subject"`
	// ExportType is the type of export.
	ExportType ExportType `json:"exportType"`
}

// NatsActivationStatus defines the observed state of NatsActivation
type NatsActivationStatus struct {
	// JWT is the JWT for the user
	JWT string `json:"jwt,omitempty"`
	// Conditions is an array of conditions that the operator is currently in.
	Conditions []metav1.Condition `json:"conditions,omitempty" optional:"true"`
	// Phase is the current phase of the operator.
	//
	// +kubebuilder:validation:Enum={None,Pending,Creating,Synchronized,Failed}
	Phase ActivationPhase `json:"phase"`
	// ControlPaused is a flag that indicates if the operator is paused.
	ControlPaused bool `json:"controlPaused,omitempty" optional:"true"`
	// LastUpdate is the timestamp of the last update.
	LastUpdate metav1.Time `json:"lastUpdate,omitempty"`
}

// +genclient
// +genreconciler
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type NatsActivation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NatsActivationSpec   `json:"spec,omitempty"`
	Status NatsActivationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NatsActivationList contains a list of NatsActivation
type NatsActivationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NatsActivation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NatsActivation{}, &NatsActivationList{})
}
