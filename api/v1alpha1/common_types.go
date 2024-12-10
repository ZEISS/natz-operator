package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	ConditionTypeSynchronizing = "Sychronizing"
	ConditionTypeSynchronized  = "Synchronized"
	ConditionTypeFailed        = "Failed"
)

const (
	ConditionReasonCreated      = "Created"
	ConditionReasonSynchronized = "Synchronized"
	ConditionReasonFailed       = "Failed"
)

const (
	FinalizerName   = "natz.zeiss.com/finalizer"
	OwnerAnnotation = "natz.zeiss.com/owner"
)

const (
	SecretPrivateKeyName = "natz.zeiss.com/nats-private-key"
	SecretSigningKeyName = "natz.zeiss.com/nats-signing-key"
)

// SecretValueFromSource represents the source of a secret value
type SecretValueFromSource struct {
	// The Secret key to select from.
	SecretKeyRef *corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}
