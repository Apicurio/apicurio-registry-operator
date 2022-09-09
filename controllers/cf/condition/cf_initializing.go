package condition

import (
	c "github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/services"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	core "k8s.io/api/core/v1"
	"net/http"
	"os"
	"time"
)

var _ loop.ControlFunction = &InitializingCF{}

type InitializingCF struct {
	ctx          context.LoopContext
	services     services.LoopServices
	httpClient   http.Client
	initializing bool

	targetType core.ServiceType
	targetIP   string

	requestOk bool
}

func NewInitializingCF(ctx context.LoopContext, services services.LoopServices) loop.ControlFunction {
	return &InitializingCF{
		ctx:      ctx,
		services: services,
		httpClient: http.Client{
			Timeout: 3 * time.Second,
		},
		initializing: true,
		requestOk:    false,
	}
}

func (this *InitializingCF) Describe() string {
	return "InitializingCF"
}

func (this *InitializingCF) Sense() {
	// This CF runs only at the initialization
	// Improve speed by avoiding unnecessary HTTP requests
	if !this.initializing || this.ctx.GetAttempts() > 0 {
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
		url := "http://" + this.targetIP + ":8080"
		res, err := this.httpClient.Get(url)
		if err == nil {
			defer res.Body.Close()
			if res.StatusCode >= 200 && res.StatusCode < 300 {
				this.requestOk = true
			} else {
				this.ctx.GetLog().V(c.V_IMPORTANT).Info("request has failed with a status", "url", url, "status", res.StatusCode)
			}
		} else if os.IsTimeout(err) {
			this.ctx.GetLog().V(c.V_IMPORTANT).Info("request has timed out", "url", url, "timeout", this.httpClient.Timeout)
		} else {
			this.ctx.GetLog().V(c.V_IMPORTANT).Info("request has failed", "url", url, "error", err.Error())
		}
	}

	if this.ctx.GetTestingSupport().IsEnabled() {
		this.requestOk = this.ctx.GetTestingSupport().GetMockCanMakeHTTPRequestToOperand()
	}
}

func (this *InitializingCF) Compare() bool {
	// Executing only when initializing
	// Prevent loop from getting stable by only executing once
	return this.initializing && this.ctx.GetAttempts() == 0
}

func (this *InitializingCF) Respond() {
	if !this.requestOk {
		this.services.GetConditionManager().GetReadyCondition().TransitionInitializing()
		this.ctx.SetRequeueDelaySoon()
	} else {
		this.initializing = false
		this.httpClient.CloseIdleConnections()
		// The condition is reset automatically
	}

	if this.ctx.GetTestingSupport().IsEnabled() {
		this.ctx.SetRequeueNow() // Ensure the reconciler is executed again very soon
	}
}

func (this *InitializingCF) Cleanup() bool {
	// No cleanup
	return true
}
