package conditions

type ApplicationNotHealthyCondition struct {
	condition
}

var _ Condition = &ApplicationNotHealthyCondition{}

func NewApplicationNotHealthyCondition() *ApplicationNotHealthyCondition {
	this := &ApplicationNotHealthyCondition{}
	this.SetType(CONDITION_TYPE_APPLICATION_NOT_HEALTHY)
	this.Reset()
	return this
}

func (this *ApplicationNotHealthyCondition) IsActive() bool {
	return this.data.Status == string(CONDITION_STATUS_TRUE)
}

// Transitions in decreasing order of priority

func (this *ApplicationNotHealthyCondition) TransitionNotReady() {
	this.data.Status = string(CONDITION_STATUS_TRUE)
	this.data.Reason = string(APPLICATION_NOT_HEALTHY_REASON_READINESS)
	this.data.Message = "Readiness probe is failing. Please check application logs."
}

func (this *ApplicationNotHealthyCondition) TransitionNotLive() {
	if this.data.Reason != string(APPLICATION_NOT_HEALTHY_REASON_READINESS) {
		this.data.Status = string(CONDITION_STATUS_TRUE)
		this.data.Reason = string(APPLICATION_NOT_HEALTHY_REASON_LIVENESS)
		this.data.Message = "Liveness probe is failing. Please check application logs."
	}
}

func (this *ApplicationNotHealthyCondition) TransitionHealthy() {
	if this.data.Reason != string(APPLICATION_NOT_HEALTHY_REASON_READINESS) &&
		this.data.Reason != string(APPLICATION_NOT_HEALTHY_REASON_LIVENESS) {
		this.data.Status = string(CONDITION_STATUS_FALSE)
		this.data.Reason = "" // The condition will be inactive
		this.data.Message = ""
	}
}
