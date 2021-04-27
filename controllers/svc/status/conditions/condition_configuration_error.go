package conditions

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigurationErrorCondition struct {
	condition
}

var _ Condition = &ConfigurationErrorCondition{}

func NewConfigurationErrorCondition() *ConfigurationErrorCondition {
	this := &ConfigurationErrorCondition{}
	this.SetType(CONDITION_TYPE_CONFIGURATION_ERROR)
	this.Reset()
	return this
}

func (this *ConfigurationErrorCondition) IsActive() bool {
	return this.data.Status == metav1.ConditionTrue
}

// Transitions in decreasing order of priority

func (this *ConfigurationErrorCondition) TransitionInvalidPersistence(currentValue string) {
	this.data.Status = metav1.ConditionTrue
	this.data.Reason = string(CONFIGURATION_ERROR_CONDITION_REASON_INVALID_PERSISTENCE)
	this.data.Message = "Invalid persistence option " + currentValue + ". Supported: <none> (or mem), sql, kafkasql."
}

func (this *ConfigurationErrorCondition) TransitionRequired(optionPath string) {
	if this.data.Reason != string(CONFIGURATION_ERROR_CONDITION_REASON_INVALID_PERSISTENCE) {

		this.data.Status = metav1.ConditionTrue
		this.data.Reason = string(CONFIGURATION_ERROR_CONDITION_REASON_REQUIRED)
		this.data.Message = "Required configuration option missing: " + optionPath + " ."
	}
}

func (this *ConfigurationErrorCondition) TransitionInvalid(currentValue string, optionPath string) {
	if this.data.Reason != string(CONFIGURATION_ERROR_CONDITION_REASON_INVALID_PERSISTENCE) &&
		this.data.Reason != string(CONFIGURATION_ERROR_CONDITION_REASON_REQUIRED) {

		this.data.Status = metav1.ConditionTrue
		this.data.Reason = string(CONFIGURATION_ERROR_CONDITION_REASON_INVALID)
		this.data.Message = "Invalid value for configuration option " + optionPath + ": " + currentValue + " ."
	}
}
