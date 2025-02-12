package v1alpha1

import (
	"github.com/zeiss/pkg/utilx"
	corev1 "k8s.io/api/core/v1"
)

const (
	// AnnotationDeletePolicy is the annotation key for the delete policy
	AnnotationDeletePolicy = "natz.zeiss.com/delete-policy"
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
	FinalizerName              = "natz.zeiss.com/finalizer"
	AccountServerFinalizerName = "natz.zeiss.com/account-server-finalizer"
	OwnerAnnotation            = "natz.zeiss.com/owner"
)

type OperationPhase string

const (
	OperationSynchronized OperationPhase = "Synchronized"
	OperationTerminating  OperationPhase = "Terminating"
	OperationFailed       OperationPhase = "Failed"
	OperationError        OperationPhase = "Error"
	OperationSucceeded    OperationPhase = "Succeeded"
)

func (os OperationPhase) Completed() bool {
	return utilx.Or(os == OperationFailed, os == OperationSucceeded)
}

func (os OperationPhase) Synchronized() bool {
	return os == OperationSynchronized
}

func (os OperationPhase) Successful() bool {
	return os == OperationSucceeded
}

func (os OperationPhase) Failed() bool {
	return os == OperationFailed
}

const (
	SecretNameKey             = "natz.zeiss.com/nats-key"
	SecretUserCredentialsName = "natz.zeiss.com/nats-user-credentials"
	SecretConfigKey           = "natz.zeiss.com/nats-config"
)

// SecretValueFromSource represents the source of a secret value
type SecretValueFromSource struct {
	// The Secret key to select from.
	SecretKeyRef *corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}
