package conditions

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ReadyCondition struct {
	condition
}

var _ Condition = &ReadyCondition{}

func NewReadyCondition() *ReadyCondition {
	this := &ReadyCondition{}
	this.SetType(CONDITION_TYPE_READY)
	this.Reset()
	return this
}

func (this *ReadyCondition) IsActive() bool {
	return true
}

// Transitions in decreasing order of priority

func (this *ReadyCondition) TransitionError() {
	this.data.Status = metav1.ConditionFalse
	this.data.Reason = string(READY_CONDITION_REASON_ERROR)
	this.data.Message = "An error occurred in the operator or the application. Please check other conditions and logs."
}

func (this *ReadyCondition) TransitionInitializing() {
	if this.data.Reason != string(READY_CONDITION_REASON_ERROR) {

		this.data.Status = metav1.ConditionFalse
		this.data.Reason = string(READY_CONDITION_REASON_INITIALIZING)
	}
}

func (this *ReadyCondition) TransitionReconciling() {
	if this.data.Reason != string(READY_CONDITION_REASON_ERROR) &&
		this.data.Reason != string(READY_CONDITION_REASON_INITIALIZING) {

		this.data.Status = metav1.ConditionFalse
		this.data.Reason = string(READY_CONDITION_REASON_RECONCILING)
	}
}

func (this *ReadyCondition) TransitionReconciled() {
	if this.data.Reason != string(READY_CONDITION_REASON_ERROR) &&
		this.data.Reason != string(READY_CONDITION_REASON_INITIALIZING) &&
		this.data.Reason != string(READY_CONDITION_REASON_RECONCILING) {

		this.data.Status = metav1.ConditionTrue
		this.data.Reason = string(READY_CONDITION_REASON_RECONCILED)
	}
}
