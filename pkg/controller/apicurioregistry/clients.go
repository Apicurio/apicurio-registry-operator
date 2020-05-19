// Some code in this file was adopted from https://github.com/atlasmap/atlasmap-operator
package apicurioregistry

import (
	"github.com/Masterminds/semver"
	ocp_config_client "github.com/openshift/client-go/config/clientset/versioned"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"net"
	"os"
)

type Clients struct {
	ctx        *Context
	config     *rest.Config
	kubeClient *KubeClient
	ocpClient  *OCPClient
}

func NewClients(ctx *Context) *Clients {
	this := &Clients{
		ctx: ctx,
	}
	config, err := inClusterConfig()
	if err != nil {
		fatal(ctx.GetLog(), err, "Could not configure clients.")
	}
	this.config = config
	this.kubeClient = NewKubeClient(ctx, config)
	this.ocpClient = NewOCPClient(ctx, config)
	return this
}

func inClusterConfig() (*rest.Config, error) {
	if len(os.Getenv("KUBERNETES_SERVICE_HOST")) == 0 {
		hosts, err := net.LookupHost("kubernetes.default.svc")
		if err != nil {
			panic(err)
		}
		os.Setenv("KUBERNETES_SERVICE_HOST", hosts[0])
	}
	if len(os.Getenv("KUBERNETES_SERVICE_PORT")) == 0 {
		os.Setenv("KUBERNETES_SERVICE_PORT", "443")
	}
	return rest.InClusterConfig()
}

func (this *Clients) OCP() *OCPClient {
	return this.ocpClient
}

func (this *Clients) Kube() *KubeClient {
	return this.kubeClient
}

func (this *Clients) IsOCP() (bool, error) {
	client, err := discovery.NewDiscoveryClientForConfig(this.config)
	if err != nil {
		return false, err
	}

	_, err = client.ServerResourcesForGroupVersion("route.openshift.io/v1")

	if err != nil && api_errors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (this *Clients) GetOCPVersion() *semver.Version {
	configClient, err := ocp_config_client.NewForConfig(this.config)
	if err != nil {
		log.Error(err, "Failed to create config client")
		return nil
	}

	var ocpSemVer *semver.Version
	clusterVersion, err := configClient.
		ConfigV1().
		ClusterVersions().
		Get("version", meta.GetOptions{})

	if err != nil {
		if api_errors.IsNotFound(err) {
			// default to OpenShift 3 as ClusterVersion API was introduced in OpenShift 4
			ocpSemVer, _ = semver.NewVersion("3")
		} else {
			log.Error(err, "Failed to get OCP cluster version")
			return nil
		}
	} else {
		v := clusterVersion.Status.History[0].Version

		ocpSemVer, err = semver.NewVersion(v)
		if err != nil {
			log.Error(err, "Failed to get OCP cluster version")
			return nil
		}
	}
	return ocpSemVer
}

func (this *Clients) IsOCP43Plus() bool {
	ocpSemVer := this.GetOCPVersion()
	if ocpSemVer != nil {
		constraint43, _ := semver.NewConstraint(">= 4.3")
		return constraint43.Check(ocpSemVer)
	}
	return false
}
