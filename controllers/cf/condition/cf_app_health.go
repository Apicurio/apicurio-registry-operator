package condition

import (
	"crypto/tls"
	"github.com/Apicurio/apicurio-registry-operator/controllers/cf"
	c "github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/services"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"go.uber.org/zap"
	core "k8s.io/api/core/v1"
	"net/http"
	"os"
	"strings"
	"time"
)

var _ loop.ControlFunction = &AppHealthCF{}

type AppHealthCF struct {
	ctx          context.LoopContext
	log          *zap.SugaredLogger
	services     services.LoopServices
	httpClient   http.Client
	initializing bool

	targetType core.ServiceType
	targetIP   string

	requestReadinessOk bool
	requestLivenessOk  bool

	disabled bool
}

func NewAppHealthCF(ctx context.LoopContext, services services.LoopServices) loop.ControlFunction {
	res := &AppHealthCF{
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
	res.log = ctx.GetLog().Sugar().With("cf", res.Describe())
	return res
}

func (this *AppHealthCF) Describe() string {
	return "AppHealthCF"
}

func (this *AppHealthCF) Sense() {
	// Improve speed by avoiding unnecessary HTTP requests
	if this.ctx.GetAttempts() > 0 {
		return
	}

	this.disabled = false
	if entry, exists := this.ctx.GetEnvCache().Get(cf.ENV_REGISTRY_AUTH_ENABLED); exists {
		this.disabled = strings.ToLower(entry.GetValue().Value) == "true"
	}
	if !this.disabled {

		var port = "8080"
		var scheme = "http://"
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
					this.log.Warnw("request to check Apicurio Registry instance readiness has failed with a status", "url", url, "status", res.StatusCode)
				}
			} else if os.IsTimeout(err) {
				this.log.Warnw("request to check Apicurio Registry instance readiness has timed out", "url", url, "timeout", this.httpClient.Timeout)
			} else {
				this.log.Warnw("request to check Apicurio Registry instance readiness has failed", "url", url)
			}
			url = scheme + this.targetIP + ":" + port + "/health/live"
			res, err = this.httpClient.Get(url)
			if err == nil {
				defer res.Body.Close()
				if res.StatusCode == 200 {
					this.requestLivenessOk = true
				} else {
					this.log.Warnw("request to check Apicurio Registry instance liveness has failed with a status", "url", url, "status", res.StatusCode)
				}
			} else if os.IsTimeout(err) {
				this.log.Warnw("request to check Apicurio Registry instance liveness has timed out", "url", url, "timeout", this.httpClient.Timeout)
			} else {
				this.log.Warnw("request to check Apicurio Registry instance liveness has failed", "url", url)
			}
		}

		if this.ctx.GetTestingSupport().IsEnabled() {
			if this.ctx.GetTestingSupport().GetMockCanMakeHTTPRequestToOperand(this.ctx.GetAppNamespace().Str()) {
				this.initializing = false
			}
			this.requestLivenessOk = this.ctx.GetTestingSupport().GetMockOperandMetricsReportReady(this.ctx.GetAppNamespace().Str())
			this.requestReadinessOk = this.ctx.GetTestingSupport().GetMockOperandMetricsReportReady(this.ctx.GetAppNamespace().Str())
		}

	} else {
		this.log.Infow("health check is disabled because auth is enabled")
		this.initializing = false
	}
}

func (this *AppHealthCF) Compare() bool {
	// Executing AFTER initialization,
	// that part is handled by InitializingCF
	// Prevent loop from getting stable by only executing once
	return !this.disabled && !this.initializing && this.ctx.GetAttempts() == 0
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

	if this.ctx.GetTestingSupport().IsEnabled() && !this.ctx.GetTestingSupport().GetMockOperandMetricsReportReady(this.ctx.GetAppNamespace().Str()) {
		this.ctx.SetRequeueNow() // Ensure the reconciler is executed again very soon
	}
}

func (this *AppHealthCF) Cleanup() bool {
	// No cleanup
	return true
}
