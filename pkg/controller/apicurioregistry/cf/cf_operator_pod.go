package cf

import (
	"os"

	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/common"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/services"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/client"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ loop.ControlFunction = &OperatorPodCF{}

type OperatorPodCF struct {
	ctx              *context.LoopContext
	svcResourceCache resources.ResourceCache
	svcClients       *client.Clients
	podExists        bool
}

// Read the operator pod into the resource cache
func NewOperatorPodCF(ctx *context.LoopContext, services *services.LoopServices) loop.ControlFunction {
	return &OperatorPodCF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),
		svcClients:       services.Clients,
		podExists:        false,
	}
}

func (this *OperatorPodCF) Describe() string {
	return "OperatorPodCF"
}

func (this *OperatorPodCF) Sense() {
	// Observation #1
	_, this.podExists = this.svcResourceCache.Get(resources.RC_KEY_OPERATOR_POD)
}

func (this *OperatorPodCF) Compare() bool {
	// Condition #1
	return !this.podExists
}

func (this *OperatorPodCF) Respond() {
	namespace := os.Getenv("POD_NAMESPACE")
	name := os.Getenv("POD_NAME")

	if namespace == "" || name == "" {
		panic("Operator could not determine name and namespace of its own pod. " +
			"Make sure that both environment variables POD_NAMESPACE and POD_NAME are present in the operators Deployment.")
	}

	// Response #1
	pod, err := this.svcClients.Kube().GetPod(common.Namespace(namespace), common.Name(name), &meta.GetOptions{})
	if err == nil && pod.GetObjectMeta().GetDeletionTimestamp() == nil {
		this.svcResourceCache.Set(resources.RC_KEY_OPERATOR_POD, resources.NewResourceCacheEntry(common.Name(name), pod))
	} else {
		this.ctx.GetLog().WithValues("type", "Warning", "error", err).
			Info("Could not read operator's Pod resource. Will retry.")
	}
}

func (this *OperatorPodCF) Cleanup() bool {
	// No cleanup
	return true
}
