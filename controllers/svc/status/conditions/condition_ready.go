package conditions

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
	this.data.Status = string(CONDITION_STATUS_FALSE)
	this.data.Reason = string(READY_CONDITION_REASON_ERROR)
	this.data.Message = "An error occurred in the operator or the application. Please check other conditions and logs."
}

func (this *ReadyCondition) TransitionInitializing() {
	if this.data.Reason != string(READY_CONDITION_REASON_ERROR) {

		this.data.Status = string(CONDITION_STATUS_FALSE)
		this.data.Reason = string(READY_CONDITION_REASON_INITIALIZING)
	}
}

func (this *ReadyCondition) TransitionReconciling() {
	if this.data.Reason != string(READY_CONDITION_REASON_ERROR) &&
		this.data.Reason != string(READY_CONDITION_REASON_INITIALIZING) {

		this.data.Status = string(CONDITION_STATUS_FALSE)
		this.data.Reason = string(READY_CONDITION_REASON_RECONCILING)
	}
}

func (this *ReadyCondition) TransitionReconciled() {
	if this.data.Reason != string(READY_CONDITION_REASON_ERROR) &&
		this.data.Reason != string(READY_CONDITION_REASON_INITIALIZING) &&
		this.data.Reason != string(READY_CONDITION_REASON_RECONCILING) {

		this.data.Status = string(CONDITION_STATUS_TRUE)
		this.data.Reason = string(READY_CONDITION_REASON_RECONCILED)
	}
}
