package conditions

import (
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type condition struct {
	previousData *metav1.Condition
	data         *metav1.Condition
	ctype        ConditionType
}

func (this *condition) SetType(ctype ConditionType) {
	this.ctype = ctype
}

func (this *condition) GetType() ConditionType {
	return this.ctype
}

func (this *condition) GetPreviousData() *metav1.Condition {
	return this.previousData
}

func (this *condition) GetData() *metav1.Condition {
	return this.data
}

func (this *condition) Reset() {
	if this.ctype == "" {
		panic("Condition type is empty.")
	}
	if this.data == nil {
		this.data = &metav1.Condition{
			Type:   string(this.GetType()),
			Status: metav1.ConditionUnknown,
		}
	}
	this.previousData = this.data
	this.data = &metav1.Condition{
		Type:               string(this.GetType()),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: this.previousData.LastTransitionTime,
	}
}

type conditionManager struct {
	conditionMap map[ConditionType]Condition
	ctx          context.LoopContext
}

var _ ConditionManager = &conditionManager{}

func NewConditionManager(ctx context.LoopContext) ConditionManager {
	this := &conditionManager{
		conditionMap: make(map[ConditionType]Condition, 3),
		ctx:          ctx,
	}
	this.conditionMap[CONDITION_TYPE_READY] = NewReadyCondition()
	this.conditionMap[CONDITION_TYPE_CONFIGURATION_ERROR] = NewConfigurationErrorCondition()
	this.conditionMap[CONDITION_TYPE_APPLICATION_NOT_HEALTHY] = NewApplicationNotHealthyCondition()
	return this
}

func (this *conditionManager) GetReadyCondition() *ReadyCondition {
	return this.conditionMap[CONDITION_TYPE_READY].(*ReadyCondition)
}

func (this *conditionManager) GetConfigurationErrorCondition() *ConfigurationErrorCondition {
	return this.conditionMap[CONDITION_TYPE_CONFIGURATION_ERROR].(*ConfigurationErrorCondition)
}

func (this *conditionManager) GetApplicationNotHealthyCondition() *ApplicationNotHealthyCondition {
	return this.conditionMap[CONDITION_TYPE_APPLICATION_NOT_HEALTHY].(*ApplicationNotHealthyCondition)
}

// Mark the status as `Reconciling` if there was a CF execution, (and reschedule) otherwise
// mask as `Reconciled`
func (this *conditionManager) AfterLoop() {
	// Error & Initializing conditions have a higher priority
	if this.ctx.GetAttempts() > 1 { // Must be 1 because some CFs always execute (AppHealthCF)
		this.GetReadyCondition().TransitionReconciling()
		// Requeue so we can try to reset the status to `Reconciled`
		this.ctx.SetRequeueDelaySoon()
	} else {
		this.GetReadyCondition().TransitionReconciled()
	}
}

func (this *conditionManager) Execute() []metav1.Condition {
	res := make([]metav1.Condition, 0)
	// TODO Would consistent ordering help performance?
	for _, v := range this.conditionMap {
		if v.IsActive() {
			previousData := v.GetPreviousData()
			data := v.GetData()
			if data.LastTransitionTime.IsZero() ||
				data.Status != previousData.Status ||
				data.Reason != previousData.Reason ||
				data.Message != previousData.Message {
				// Update time if the condition changed
				data.LastTransitionTime = metav1.Now()
			}
			res = append(res, *data)
		}
		v.Reset()
	}
	return res
}
