package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GatewayPhase string

const (
	GatewayPhaseNone     GatewayPhase = ""
	GatewayPhaseCreating GatewayPhase = "Creating"
	GatewayPhaseActive   GatewayPhase = "Active"
	GatewaySynchronized  GatewayPhase = "Synchronized"
	GatewayPhaseFailed   GatewayPhase = "Failed"
)

// NatsGatewayRef is a reference to a NatsGateway
type NatsgatewayReference struct {
	// Name is the name of the gateway
	Name string `json:"name"`
	// Namespace is the namespace of the gateway
	Namespace string `json:"namespace"`
}

type NatsGatewaySpec struct {
	// URL is the URL of the gateway.
	URL string `json:"url"`
	// Username is the username of the gateway.
	Username SecretValueFromSource `json:"username,omitempty"`
	// Password is the password of the gateway.
	Password SecretValueFromSource `json:"password,omitempty"`
}

type NatsGatewayStatus struct {
	// Conditions is an array of conditions that the operator is currently in.
	Conditions []metav1.Condition `json:"conditions,omitempty" optional:"true"`
	// Phase is the current phase of the operator.
	//
	// +kubebuilder:validation:Enum={None,Pending,Creating,Synchronized,Failed}
	Phase GatewayPhase `json:"phase"`
	// ControlPaused is a flag that indicates if the operator is paused.
	ControlPaused bool `json:"controlPaused,omitempty" optional:"true"`
	// LastUpdate is the timestamp of the last update.
	LastUpdate metav1.Time `json:"lastUpdate,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +genreconciler

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
