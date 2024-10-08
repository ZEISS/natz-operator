package v1alpha1

import (
	"github.com/nats-io/jwt/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NatsOperatorSpec struct {
	SigningKeys jwt.StringList `json:"signing_keys,omitempty"`
}
type NatsOperatorStatus struct {
	OperatorSecretName string `json:"operatorSecretName,omitempty"`
	PublicKey          string `json:"publicKey,omitempty"`
	JWT                string `json:"jwt,omitempty"`
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
