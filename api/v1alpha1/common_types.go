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
	SecretNameKey             = "natz.zeiss.com/nats-key"
	SecretUserCredentialsName = "natz.zeiss.com/nats-user-credentials"
)

// SecretValueFromSource represents the source of a secret value
type SecretValueFromSource struct {
	// The Secret key to select from.
	SecretKeyRef *corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}
