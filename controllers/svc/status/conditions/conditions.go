package conditions

import (
	api "github.com/Apicurio/apicurio-registry-operator/api/v1"
)

type ConditionType string

const (
	CONDITION_TYPE_READY                   ConditionType = "Ready"
	CONDITION_TYPE_CONFIGURATION_ERROR     ConditionType = "ConfigurationError"
	CONDITION_TYPE_APPLICATION_NOT_HEALTHY ConditionType = "ApplicationNotHealthy"
	// CONDITION_TYPE_OPERATOR_ERROR ConditionType = "OperatorError" // General error
)

type ConditionStatus string

const (
	CONDITION_STATUS_TRUE    ConditionStatus = "True"
	CONDITION_STATUS_FALSE   ConditionStatus = "False"
	CONDITION_STATUS_UNKNOWN ConditionStatus = "Unknown"
)

type Condition interface {
	SetType(ConditionType)

	GetType() ConditionType

	// References an internal state. Make a copy before using the data unless intended.
	GetPreviousData() *api.ApicurioRegistryStatusCondition

	// References an internal state. Make a copy before using the data unless intended.
	GetData() *api.ApicurioRegistryStatusCondition

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

	Execute() []api.ApicurioRegistryStatusCondition
}
