package condition

import (
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/services"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	core "k8s.io/api/core/v1"
	"net/http"
)

var _ loop.ControlFunction = &AppHealthCF{}

type AppHealthCF struct {
	ctx      *context.LoopContext
	services *services.LoopServices

	initializing bool

	targetType core.ServiceType
	targetIP   string

	requestReadinessOk bool
	requestLivenessOk  bool
}

func NewAppHealthCF(ctx *context.LoopContext, services *services.LoopServices) loop.ControlFunction {
	return &AppHealthCF{
		ctx:      ctx,
		services: services,

		initializing:       true,
		requestReadinessOk: false,
		requestLivenessOk:  false,
	}
}

func (this *AppHealthCF) Describe() string {
	return "AppHealthCF"
}

func (this *AppHealthCF) Sense() {

	if serviceEntry, exists := this.ctx.GetResourceCache().Get(resources.RC_KEY_SERVICE); exists {
		this.targetType = serviceEntry.GetValue().(*core.Service).Spec.Type
		this.targetIP = serviceEntry.GetValue().(*core.Service).Spec.ClusterIP
	}

	this.requestReadinessOk = false
	this.requestLivenessOk = false
	if this.targetType == core.ServiceTypeClusterIP && this.targetIP != "" {
		if res, err := http.Get("http://" + this.targetIP + ":8080/health/ready"); err == nil {
			defer res.Body.Close()
			if res.StatusCode == 200 {
				this.requestReadinessOk = true
				this.initializing = false
			}
		}
		if res, err := http.Get("http://" + this.targetIP + ":8080/health/live"); err == nil {
			defer res.Body.Close()
			if res.StatusCode == 200 {
				this.requestLivenessOk = true
			}
		}
	}
}

func (this *AppHealthCF) Compare() bool {
	return !this.initializing && this.ctx.GetAttempts() == 0
}

func (this *AppHealthCF) Respond() {
	if !this.requestReadinessOk {
		this.services.GetConditionManager().GetApplicationNotHealthyCondition().TransitionNotReady()
		this.services.GetConditionManager().GetReadyCondition().TransitionError()
		this.ctx.SetRequeueDelaySoon()
		this.initializing = false
	}
	if !this.requestLivenessOk {
		this.services.GetConditionManager().GetApplicationNotHealthyCondition().TransitionNotLive()
		this.services.GetConditionManager().GetReadyCondition().TransitionError()
		this.ctx.SetRequeueDelaySoon()
		this.initializing = false
	}
	this.ctx.SetRequeueDelaySec(3 * 60) // 3 min
}

func (this *AppHealthCF) Cleanup() bool {
	// No cleanup
	return true
}
