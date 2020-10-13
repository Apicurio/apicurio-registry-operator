// Some code in this file was adopted from https://github.com/atlasmap/atlasmap-operator
package apicurioregistry

import (
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"net"
	"os"
  "os/user"
	"path"
	"github.com/Masterminds/semver"
	ocp_config_client "github.com/openshift/client-go/config/clientset/versioned"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"	
)

// RecommendedConfigPathEnvVar is a environment variable for path configuration
const RecommendedConfigPathEnvVar = "KUBECONFIG"

type Clients struct {
	ctx              loop.ControlLoopContext
	config           *rest.Config
	kubeClient       *KubeClient
	ocpClient        *OCPClient
	crdClient        *CRDClient
	monitoringClient *MonitoringClient
}

func NewClients(ctx loop.ControlLoopContext) *Clients {
	this := &Clients{
		ctx: ctx,
	}
	config, err := inClusterConfig()
	if err != nil {
		fatal(ctx.GetLog(), err, "Could not configure clients.")
	}
	this.config = config

	config, err = inClusterConfig()
	this.kubeClient = NewKubeClient(ctx, config)

	config, err = inClusterConfig()
	this.ocpClient = NewOCPClient(ctx, config)

	config, err = inClusterConfig()
	this.crdClient = NewCRDClient(ctx, config)

	config, err = inClusterConfig()
	this.monitoringClient = NewMonitoringClient(ctx, config)

	return this
}

func inClusterConfig() (*rest.Config, error) {


		if len(os.Getenv("KUBERNETES_SERVICE_HOST")) == 0 {
			hosts, err := net.LookupHost("kubernetes.default.svc")
			if err != nil {
				return outOfClusterConfig()
			}
			if err := os.Setenv("KUBERNETES_SERVICE_HOST", hosts[0]); err != nil {
				return nil, err
			}
		}
		if len(os.Getenv("KUBERNETES_SERVICE_PORT")) == 0 {
			if err := os.Setenv("KUBERNETES_SERVICE_PORT", "443"); err != nil {
				panic(err)
			}
		}

	return rest.InClusterConfig()
}

func outOfClusterConfig() (*rest.Config, error) {

	configFile := getKubeConfigFile()

	if len(configFile) > 0 {

		log.Info("Reading config from file"+  configFile)
		// use the current context in kubeconfig
		// This is very useful for running locally.
		clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: configFile},
			&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}})

		config, err := clientConfig.ClientConfig()

		return config, err

	}
	return rest.InClusterConfig()
}

//GetKubeConfigFile tries to find a kubeconfig file.
func getKubeConfigFile() string {
	configFile := ""

	usr, err := user.Current()
	if err != nil {
		log.Info("Could not get current user; error %v", err)
	} else {
		configFile = path.Join(usr.HomeDir, ".kube", "config")
	}

	if len(os.Getenv(RecommendedConfigPathEnvVar)) > 0 {
		configFile = os.Getenv(RecommendedConfigPathEnvVar)
	}

	return configFile
}

func (this *Clients) OCP() *OCPClient {
	return this.ocpClient
}

func (this *Clients) Kube() *KubeClient {
	return this.kubeClient
}

func (this *Clients) CRD() *CRDClient {
	return this.crdClient
}

func (this *Clients) Monitoring() *MonitoringClient {
	return this.monitoringClient
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
