package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigPhase string

const (
	ConfigPhaseNone     ConfigPhase = ""
	ConfigPhaseCreating ConfigPhase = "Creating"
	ConfigPhaseActive   ConfigPhase = "Active"
	ConfigPhaseFailed   ConfigPhase = "Failed"
)

// NatsConfigSpec defines the desired state of NatsConfig
type NatsConfigSpec struct{}

// NatsConfigStatus defines the observed state of NatsConfig
type NatsConfigStatus struct {
	// Conditions is an array of conditions that the operator is currently in.
	Conditions []metav1.Condition `json:"conditions,omitempty" optional:"true"`
	// Phase is the current phase of the operator.
	//
	// +kubebuilder:validation:Enum={None,Pending,Creating,Synchronized,Failed}
	Phase ConfigPhase `json:"phase"`
	// ControlPaused is a flag that indicates if the operator is paused.
	ControlPaused bool `json:"controlPaused,omitempty" optional:"true"`
	// LastUpdate is the timestamp of the last update.
	LastUpdate metav1.Time `json:"lastUpdate,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type NatsConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NatsConfigSpec   `json:"spec,omitempty"`
	Status NatsConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NatsConfigList contains a list of NatsConfig
type NatsConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NatsConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NatsConfig{}, &NatsConfigList{})
}
