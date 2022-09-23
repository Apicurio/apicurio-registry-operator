package condition

import (
	"crypto/tls"
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

var _ loop.ControlFunction = &AppHealthCF{}

type AppHealthCF struct {
	ctx          context.LoopContext
	services     services.LoopServices
	httpClient   http.Client
	initializing bool

	targetType core.ServiceType
	targetIP   string

	requestReadinessOk bool
	requestLivenessOk  bool
}

func NewAppHealthCF(ctx context.LoopContext, services services.LoopServices) loop.ControlFunction {
	return &AppHealthCF{
		ctx:      ctx,
		services: services,
		httpClient: http.Client{
			Timeout: 3 * time.Second,
			Transport: &http.Transport{
				// ignore expired SSL certificates for health checks
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
		initializing:       true,
		requestReadinessOk: false,
		requestLivenessOk:  false,
	}
}

func (this *AppHealthCF) Describe() string {
	return "AppHealthCF"
}

func (this *AppHealthCF) Sense() {
	// Improve speed by avoiding unnecessary HTTP requests
	if this.ctx.GetAttempts() > 0 {
		return
	}

	var port string = "8080"
	var scheme string = "http://"
	if serviceEntry, exists := this.ctx.GetResourceCache().Get(resources.RC_KEY_SERVICE); exists {
		this.targetType = serviceEntry.GetValue().(*core.Service).Spec.Type
		this.targetIP = serviceEntry.GetValue().(*core.Service).Spec.ClusterIP

		if c.HasPort("https", serviceEntry.GetValue().(*core.Service).Spec.Ports) {
			port = "8443"
			scheme = "https://"
		}

	}

	this.requestReadinessOk = false
	this.requestLivenessOk = false
	if this.targetType == core.ServiceTypeClusterIP && this.targetIP != "" {
		url := scheme + this.targetIP + ":" + port + "/health/ready"
		res, err := this.httpClient.Get(url)
		if err == nil {
			// TODO Unify this with InitializingCF?
			defer res.Body.Close()
			if res.StatusCode == 200 {
				this.requestReadinessOk = true
				this.initializing = false
			} else {
				this.ctx.GetLog().V(c.V_IMPORTANT).Info("request has failed with a status", "url", url, "status", res.StatusCode)
			}
		} else if os.IsTimeout(err) {
			this.ctx.GetLog().V(c.V_IMPORTANT).Info("request has timed out", "url", url, "timeout", this.httpClient.Timeout)
		} else {
			this.ctx.GetLog().V(c.V_IMPORTANT).Info("request has failed", "url", url, "error", err.Error())
		}
		url = scheme + this.targetIP + ":" + port + "/health/live"
		res, err = this.httpClient.Get(url)
		if err == nil {
			defer res.Body.Close()
			if res.StatusCode == 200 {
				this.requestLivenessOk = true
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
		if this.ctx.GetTestingSupport().GetMockCanMakeHTTPRequestToOperand() {
			this.initializing = false
		}
		this.requestLivenessOk = this.ctx.GetTestingSupport().GetMockOperandMetricsReportReady()
		this.requestReadinessOk = this.ctx.GetTestingSupport().GetMockOperandMetricsReportReady()
	}
}

func (this *AppHealthCF) Compare() bool {
	// Executing AFTER initialization,
	// that part is handled by InitializingCF
	// Prevent loop from getting stable by only executing once
	return !this.initializing && this.ctx.GetAttempts() == 0
}

func (this *AppHealthCF) Respond() {
	if !this.requestReadinessOk {
		this.services.GetConditionManager().GetApplicationNotHealthyCondition().TransitionNotReady()
		this.services.GetConditionManager().GetReadyCondition().TransitionError()
		this.ctx.SetRequeueDelaySoon()
	}
	if !this.requestLivenessOk {
		this.services.GetConditionManager().GetApplicationNotHealthyCondition().TransitionNotLive()
		this.services.GetConditionManager().GetReadyCondition().TransitionError()
		this.ctx.SetRequeueDelaySoon()
	}

	this.ctx.SetRequeueDelaySec(3 * 60) // 3 min

	if this.ctx.GetTestingSupport().IsEnabled() && !this.ctx.GetTestingSupport().GetMockOperandMetricsReportReady() {
		this.ctx.SetRequeueNow() // Ensure the reconciler is executed again very soon
	}
}

func (this *AppHealthCF) Cleanup() bool {
	// No cleanup
	return true
}
