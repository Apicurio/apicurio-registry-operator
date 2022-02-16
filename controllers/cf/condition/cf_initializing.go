package condition

import (
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/services"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	core "k8s.io/api/core/v1"
	"net/http"
)

var _ loop.ControlFunction = &InitializingCF{}

type InitializingCF struct {
	ctx      *context.LoopContext
	services *services.LoopServices

	initializing bool

	targetType core.ServiceType
	targetIP   string

	requestOk bool
}

func NewInitializingCF(ctx *context.LoopContext, services *services.LoopServices) loop.ControlFunction {
	return &InitializingCF{
		ctx:      ctx,
		services: services,

		initializing: true,
		requestOk:    false,
	}
}

func (this *InitializingCF) Describe() string {
	return "InitializingCF"
}

func (this *InitializingCF) Sense() {
	// This CF runs only at the initialization.
	if !this.initializing {
		return
	}
	// The application is initialized if we can make an HTTP request to the app via the Service
	// (as Ingress/Route might not work on some systems, or without additional config).

	if serviceEntry, exists := this.ctx.GetResourceCache().Get(resources.RC_KEY_SERVICE); exists {
		this.targetType = serviceEntry.GetValue().(*core.Service).Spec.Type
		this.targetIP = serviceEntry.GetValue().(*core.Service).Spec.ClusterIP
	}

	this.requestOk = false
	if this.targetType == core.ServiceTypeClusterIP && this.targetIP != "" {
		client := &http.Client{
			// Do not follow redirects (for when TLS is enabled)
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		res, err := client.Get("http://" + this.targetIP + ":8080")
		if err == nil {
			if (res.StatusCode >= 200 && res.StatusCode < 300) || res.StatusCode == 301 {
				this.requestOk = true
			}
		}
	}
}

func (this *InitializingCF) Compare() bool {
	return this.initializing && this.ctx.GetAttempts() == 0
}

func (this *InitializingCF) Respond() {
	if !this.requestOk {
		this.services.GetConditionManager().GetReadyCondition().TransitionInitializing()
		this.ctx.SetRequeueDelaySoon()
	} else {
		this.initializing = false
	}
}

func (this *InitializingCF) Cleanup() bool {
	// No cleanup
	return true
}
