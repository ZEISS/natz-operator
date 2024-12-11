package status

import (
	"fmt"

	natsv1alpha1 "github.com/zeiss/natz-operator/api/v1alpha1"

	"github.com/zeiss/pkg/slices"
	"github.com/zeiss/pkg/utilx"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewCondition ...
func NewCondition(conditionType string, conditionStatus metav1.ConditionStatus, now metav1.Time, reason, message string) metav1.Condition {
	return metav1.Condition{
		Type:               conditionType,
		Status:             conditionStatus,
		LastTransitionTime: now,
		Reason:             reason,
		Message:            message,
	}
}

// SetCondition ...
func SetCondition(condition metav1.Condition, conditions ...metav1.Condition) []metav1.Condition {
	return utilx.IfElse(
		slices.Any(func(cond metav1.Condition) bool {
			return cond.Type == condition.Type
		}, conditions...),
		conditions,
		append(conditions, condition),
	)
}

// SetNatzPrivateKeyCondition ...
func SetNatzPrivateKeyCondition(obj *natsv1alpha1.NatsPrivateKey, condition metav1.Condition) {
	obj.Status.Conditions = SetCondition(condition, obj.Status.Conditions...)
}

// SetNatzOperatorCondition ...
func SetNatzOperatorCondition(obj *natsv1alpha1.NatsOperator, condition metav1.Condition) {
	obj.Status.Conditions = SetCondition(condition, obj.Status.Conditions...)
}

// SetNatzAccountCondition ...
func SetNatzAccountCondition(obj *natsv1alpha1.NatsAccount, condition metav1.Condition) {
	obj.Status.Conditions = SetCondition(condition, obj.Status.Conditions...)
}

// SetNatzSigningKeyCondition ...
func SetNatzSigningKeyCondition(obj *natsv1alpha1.NatsSigningKey, condition metav1.Condition) {
	obj.Status.Conditions = SetCondition(condition, obj.Status.Conditions...)
}

// SetNatzUserCondition ...
func SetNatzUserCondition(obj *natsv1alpha1.NatsUser, condition metav1.Condition) {
	obj.Status.Conditions = SetCondition(condition, obj.Status.Conditions...)
}

// SetNatzConfigCondition ...
func SetNatzConfigCondition(obj *natsv1alpha1.NatsConfig, condition metav1.Condition) {
	obj.Status.Conditions = SetCondition(condition, obj.Status.Conditions...)
}

// NewNatzConfigSynchronizedCondition creates the provisioning started condition in cluster conditions.
func NewNatzConfigSynchronizedCondition(obj *natsv1alpha1.NatsConfig) metav1.Condition {
	return metav1.Condition{
		Type:               natsv1alpha1.ConditionTypeSynchronized,
		ObservedGeneration: obj.Generation,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Message:            fmt.Sprintf("the config has successfully created: %s", obj.Name),
		Reason:             natsv1alpha1.ConditionReasonSynchronized,
	}
}

// NewNatzConfigFailedCondition creates the provisioning started condition in cluster conditions.
func NewNatzConfigFailedCondition(obj *natsv1alpha1.NatsConfig, err error) metav1.Condition {
	return metav1.Condition{
		Type:               natsv1alpha1.ConditionTypeFailed,
		ObservedGeneration: obj.Generation,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Message:            err.Error(),
		Reason:             natsv1alpha1.ConditionReasonFailed,
	}
}

// NewPrivateKeySychronizedCondition creates the provisioning started condition in cluster conditions.
func NewPrivateKeySychronizedCondition(obj *natsv1alpha1.NatsPrivateKey) metav1.Condition {
	return metav1.Condition{
		Type:               natsv1alpha1.ConditionTypeSynchronized,
		ObservedGeneration: obj.Generation,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Message:            fmt.Sprintf("the private key has successfully created: %s", obj.Name),
		Reason:             natsv1alpha1.ConditionReasonSynchronized,
	}
}

// NewOperatorSychronizedCondition creates the provisioning started condition in cluster conditions.
func NewOperatorSychronizedCondition(obj *natsv1alpha1.NatsOperator) metav1.Condition {
	return metav1.Condition{
		Type:               natsv1alpha1.ConditionTypeSynchronized,
		ObservedGeneration: obj.Generation,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Message:            fmt.Sprintf("the operator has successfully created: %s", obj.Name),
		Reason:             natsv1alpha1.ConditionReasonSynchronized,
	}
}

// NewOperatorFailedCondition creates the provisioning started condition in cluster conditions.
func NewOperatorFailedCondition(obj *natsv1alpha1.NatsOperator, err error) metav1.Condition {
	return metav1.Condition{
		Type:               natsv1alpha1.ConditionTypeFailed,
		ObservedGeneration: obj.Generation,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Message:            err.Error(),
		Reason:             natsv1alpha1.ConditionReasonFailed,
	}
}

// NewAccountSychronizedCondition creates the provisioning started condition in cluster conditions.
func NewAccountSychronizedCondition(obj *natsv1alpha1.NatsAccount) metav1.Condition {
	return metav1.Condition{
		Type:               natsv1alpha1.ConditionTypeSynchronized,
		ObservedGeneration: obj.Generation,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Message:            fmt.Sprintf("the account has successfully created: %s", obj.Name),
		Reason:             natsv1alpha1.ConditionReasonSynchronized,
	}
}

// NewAccountFailedCondition creates the provisioning started condition in cluster conditions.
func NewAccountFailedCondition(obj *natsv1alpha1.NatsAccount, err error) metav1.Condition {
	return metav1.Condition{
		Type:               natsv1alpha1.ConditionTypeFailed,
		ObservedGeneration: obj.Generation,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Message:            err.Error(),
		Reason:             natsv1alpha1.ConditionReasonFailed,
	}
}

// NewSigningKeySychronizedCondition creates the provisioning started condition in cluster conditions.
func NewSigningKeySychronizedCondition(obj *natsv1alpha1.NatsSigningKey) metav1.Condition {
	return metav1.Condition{
		Type:               natsv1alpha1.ConditionTypeSynchronized,
		ObservedGeneration: obj.Generation,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Message:            fmt.Sprintf("the signing key has successfully created: %s", obj.Name),
		Reason:             natsv1alpha1.ConditionReasonSynchronized,
	}
}

// NewUserSychronizedCondition creates the provisioning started condition in cluster conditions.
func NewUserSychronizedCondition(obj *natsv1alpha1.NatsUser) metav1.Condition {
	return metav1.Condition{
		Type:               natsv1alpha1.ConditionTypeSynchronized,
		ObservedGeneration: obj.Generation,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Message:            fmt.Sprintf("the user has successfully created: %s", obj.Name),
		Reason:             natsv1alpha1.ConditionReasonSynchronized,
	}
}

// NewUserFailedCondition creates the provisioning started condition in cluster conditions.
func NewUserFailedCondition(obj *natsv1alpha1.NatsUser, err error) metav1.Condition {
	return metav1.Condition{
		Type:               natsv1alpha1.ConditionTypeFailed,
		ObservedGeneration: obj.Generation,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Message:            err.Error(),
		Reason:             natsv1alpha1.ConditionReasonFailed,
	}
}

// NewSigningKeyFailedCondition creates the provisioning started condition in cluster conditions.
func NewSigningKeyFailedCondition(obj *natsv1alpha1.NatsSigningKey, err error) metav1.Condition {
	return metav1.Condition{
		Type:               natsv1alpha1.ConditionTypeFailed,
		ObservedGeneration: obj.Generation,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Message:            err.Error(),
		Reason:             natsv1alpha1.ConditionReasonFailed,
	}
}

// NewPrivateKeyFailedCondition creates the provisioning started condition in cluster conditions.
func NewPrivateKeyFailedCondition(obj *natsv1alpha1.NatsPrivateKey, err error) metav1.Condition {
	return metav1.Condition{
		Type:               natsv1alpha1.ConditionTypeFailed,
		ObservedGeneration: obj.Generation,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Message:            err.Error(),
		Reason:             natsv1alpha1.ConditionReasonFailed,
	}
}
