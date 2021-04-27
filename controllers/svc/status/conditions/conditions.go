package conditions

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConditionType string

const (
	CONDITION_TYPE_READY                   ConditionType = "Ready"
	CONDITION_TYPE_CONFIGURATION_ERROR     ConditionType = "ConfigurationError"
	CONDITION_TYPE_APPLICATION_NOT_HEALTHY ConditionType = "ApplicationNotHealthy"
	// CONDITION_TYPE_OPERATOR_ERROR ConditionType = "OperatorError" // General error
)

type Condition interface {
	SetType(ConditionType)

	GetType() ConditionType

	// References an internal state. Make a copy before using the data unless intended.
	GetPreviousData() *metav1.Condition

	// References an internal state. Make a copy before using the data unless intended.
	GetData() *metav1.Condition

	IsActive() bool

	Reset()
}

// ========== ReadyCondition ==========

type ReadyConditionReason string

const (
	// Priority ordered
	READY_CONDITION_REASON_ERROR        ReadyConditionReason = "Error"
	READY_CONDITION_REASON_INITIALIZING ReadyConditionReason = "Initializing"
	READY_CONDITION_REASON_RECONCILING  ReadyConditionReason = "Reconciling"
	READY_CONDITION_REASON_RECONCILED   ReadyConditionReason = "Reconciled"
)

// ========== ConfigurationErrorCondition ==========

type ConfigurationErrorConditionReason string

const (
	// Priority ordered
	CONFIGURATION_ERROR_CONDITION_REASON_INVALID_PERSISTENCE ConfigurationErrorConditionReason = "InvalidPersistenceOption"
	CONFIGURATION_ERROR_CONDITION_REASON_REQUIRED            ConfigurationErrorConditionReason = "MissingRequiredOption"
	CONFIGURATION_ERROR_CONDITION_REASON_INVALID             ConfigurationErrorConditionReason = "InvalidValue"
)

// ========== ApplicationNotHealthyCondition ==========

type ApplicationNotHealthyConditionReason string

const (
	// Priority ordered
	APPLICATION_NOT_HEALTHY_REASON_READINESS ApplicationNotHealthyConditionReason = "ReadinessProbeFailed"
	APPLICATION_NOT_HEALTHY_REASON_LIVENESS  ApplicationNotHealthyConditionReason = "LivenessProbeFailed"
)

// ========== ConditionManager ==========

type ConditionManager interface {
	GetReadyCondition() *ReadyCondition

	GetConfigurationErrorCondition() *ConfigurationErrorCondition

	GetApplicationNotHealthyCondition() *ApplicationNotHealthyCondition

	// Run after the control loop
	AfterLoop()

	Execute() []metav1.Condition
}
