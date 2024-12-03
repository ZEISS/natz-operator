package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GatewayPhase string

const (
	GatewayPhaseNone     GatewayPhase = ""
	GatewayPhaseCreating GatewayPhase = "Creating"
	GatewayPhaseActive   GatewayPhase = "Active"
	GatewayPhaseFailed   GatewayPhase = "Failed"
)

type NatsGatewaySpec struct {
	URL      string                `json:"url"`
	Name     string                `json:"name,omitempty"`
	Username string                `json:"username"`
	Password SecretValueFromSource `json:"password"`
}

type NatsGatewayStatus struct {
	// Phase is the current state of the gateway
	Phase GatewayPhase `json:"phase"`

	// ControlPaused indicates if the controller paused the control of the gateway
	ControlPaused bool `json:"controlPaused,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type NatsGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NatsGatewaySpec   `json:"spec,omitempty"`
	Status NatsGatewayStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NatsGatewayList contains a list of NatsGateway
type NatsGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NatsGateway `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NatsGateway{}, &NatsGatewayList{})
}
