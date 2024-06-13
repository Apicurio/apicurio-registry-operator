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

var _ loop.ControlFunction = &InitializingCF{}

type InitializingCF struct {
	ctx          context.LoopContext
	log          *zap.SugaredLogger
	services     services.LoopServices
	httpClient   http.Client
	initializing bool

	targetType core.ServiceType
	targetIP   string

	requestOk bool

	disabled bool
}

func NewInitializingCF(ctx context.LoopContext, services services.LoopServices) loop.ControlFunction {
	res := &InitializingCF{
		ctx:      ctx,
		services: services,
		httpClient: http.Client{
			Timeout: 3 * time.Second,
			Transport: &http.Transport{
				// ignore expired SSL certificates for health checks
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
		initializing: true,
		requestOk:    false,
	}
	res.log = ctx.GetLog().Sugar().With("cf", res.Describe())
	return res
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

	this.disabled = false
	if entry, exists := this.ctx.GetEnvCache().Get(cf.ENV_REGISTRY_AUTH_ENABLED); exists {
		this.disabled = strings.ToLower(entry.GetValue().Value) == "true"
	}
	if !this.disabled {

		// The application is initialized if we can make an HTTP request to the app via the Service
		// (as Ingress/Route might not work on some systems, or without additional config).

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

		this.requestOk = false
		if this.ctx.GetTestingSupport().IsEnabled() {
			this.requestOk = this.ctx.GetTestingSupport().GetMockCanMakeHTTPRequestToOperand(this.ctx.GetAppNamespace().Str())
		} else {
			if this.targetType == core.ServiceTypeClusterIP && this.targetIP != "" {
				// NOTE: The client will follow redirects, but I have found that there is a strange issue with a cyclic redirect:
				// http://172.30.162.200:8080 -> http://172.30.162.200:8080/ui -> http://172.30.162.200:8080/ui
				// that ends with the client returning status 404. Therefore, we are using /apis instead.
				url := scheme + this.targetIP + ":" + port + "/apis"
				res, err := this.httpClient.Get(url)
				if err == nil {
					defer res.Body.Close()
					if res.StatusCode >= 200 && res.StatusCode < 300 {
						this.requestOk = true
					} else {
						this.log.Warnw("request to check that Apicurio Registry instance is available has failed with a status", "url", url, "status", res.StatusCode)
					}
				} else if os.IsTimeout(err) {
					this.log.Warnw("request to check that Apicurio Registry instance is available has timed out", "url", url, "timeout", this.httpClient.Timeout, "err", err)
				} else {
					this.log.Warnw("request to check that Apicurio Registry instance is available has failed", "url", url, "err", err)
				}
			}
		}

	} else {
		this.log.Infow("initializing health check is disabled because auth is enabled")
		this.initializing = false
	}
}

func (this *InitializingCF) Compare() bool {
	// Executing only when initializing
	// Prevent loop from getting stable by only executing once
	this.log.Debugln("this.disabled", this.disabled)
	this.log.Debugln("this.initializing", this.initializing)
	this.log.Debugln("this.ctx.GetAttempts()", this.ctx.GetAttempts())
	return !this.disabled && this.initializing && this.ctx.GetAttempts() == 0
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
