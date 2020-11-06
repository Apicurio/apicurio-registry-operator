// Some code in this file was adopted from https://github.com/atlasmap/atlasmap-operator
package client

import (
	"net"
	"os"
	"os/user"
	"path"

	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/common"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// RecommendedConfigPathEnvVar is a environment variable for path configuration
const RecommendedConfigPathEnvVar = "KUBECONFIG"

var isOpenshift *bool

var log = logf.Log.WithName("controller_apicurioregistry-Clients")

type Clients struct {
	ctx              *context.LoopContext
	config           *rest.Config
	kubeClient       *KubeClient
	ocpClient        *OCPClient
	crdClient        *CRDClient
	monitoringClient *MonitoringClient
}

func NewClients(ctx *context.LoopContext) *Clients {
	this := &Clients{
		ctx: ctx,
	}
	config, err := inClusterConfig()
	if err != nil {
		common.Fatal(ctx.GetLog(), err, "Could not configure clients.")
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

		log.Info("Reading config from file" + configFile)
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

func IsOCP() (bool, error) {
	if isOpenshift == nil {
		o, err := detectOpenshift()
		if err != nil {
			return o, err
		}
		isOpenshift = &o
	}
	return *isOpenshift, nil
}

func detectOpenshift() (bool, error) {
	config, err := inClusterConfig()
	if err != nil {
		return false, err
	}

	client, err := discovery.NewDiscoveryClientForConfig(config)
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

func getSpec(ctx *context.LoopContext) *ar.ApicurioRegistry {
	entry, exists := ctx.GetResourceCache().Get(resources.RC_KEY_SPEC)
	if !exists {
		panic("Could not get ApicurioRegistry from resource cache.")
	}
	return entry.GetValue().(*ar.ApicurioRegistry)
}
