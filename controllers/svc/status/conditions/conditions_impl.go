package conditions

import (
	api "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"time"
)

type condition struct {
	previousData *api.ApicurioRegistryStatusCondition
	data         *api.ApicurioRegistryStatusCondition
	ctype        ConditionType
}

func (this *condition) SetType(ctype ConditionType) {
	this.ctype = ctype
}

func (this *condition) GetType() ConditionType {
	return this.ctype
}

func (this *condition) GetPreviousData() *api.ApicurioRegistryStatusCondition {
	return this.previousData
}

func (this *condition) GetData() *api.ApicurioRegistryStatusCondition {
	return this.data
}

func (this *condition) Reset() {
	if this.ctype == "" {
		panic("Condition type not set!")
	}
	if this.data == nil {
		this.data = &api.ApicurioRegistryStatusCondition{
			Type:   string(this.GetType()),
			Status: string(CONDITION_STATUS_UNKNOWN),
		}
	}
	this.previousData = this.data
	this.data = &api.ApicurioRegistryStatusCondition{
		Type:               string(this.GetType()),
		Status:             string(CONDITION_STATUS_UNKNOWN),
		LastTransitionTime: this.previousData.LastTransitionTime,
	}
}

type conditionManager struct {
	conditionMap map[ConditionType]Condition
	ctx          *context.LoopContext
}

var _ ConditionManager = &conditionManager{}

func NewConditionManager(ctx *context.LoopContext) ConditionManager {
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
	// TODO Make this not `ReadyCondition` specific (down one level)
	this.GetReadyCondition().TransitionReconciled()
	if this.ctx.GetAttempts() > 1 { // Must be 1 because some CFs always execute (AppHealthCF)
		this.GetReadyCondition().TransitionReconciling()
		this.ctx.SetRequeueDelaySoon()
	}
}

func (this *conditionManager) Execute() []api.ApicurioRegistryStatusCondition {
	res := make([]api.ApicurioRegistryStatusCondition, 0)
	now := time.Now().UTC().Format(time.RFC3339)
	for _, v := range this.conditionMap {
		if v.IsActive() {
			previousData := v.GetPreviousData()
			data := v.GetData()
			if data.LastTransitionTime == "" ||
				data.Status != previousData.Status ||
				data.Reason != previousData.Reason ||
				data.Message != previousData.Message {
				// Update time if the condition changed
				data.LastTransitionTime = now
			}
			res = append(res, *data)
		}
		v.Reset()
	}
	return res
}
