package apicurioregistry

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
)

var _ ControlFunction = &OperatorPodCF{}

type OperatorPodCF struct {
	ctx       *Context
	podExists bool
}

// Read the operator pod into the resource cache
func NewOperatorPodCF(ctx *Context) ControlFunction {
	return &OperatorPodCF{
		ctx:       ctx,
		podExists: false,
	}
}

func (this *OperatorPodCF) Describe() string {
	return "OperatorPodCF"
}

func (this *OperatorPodCF) Sense() {
	// Observation #1
	_, this.podExists = this.ctx.GetResourceCache().Get(RC_KEY_OPERATOR_POD)
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
	pod, err := this.ctx.GetClients().Kube().GetPod(namespace, name, &meta.GetOptions{})
	if err == nil && pod.GetObjectMeta().GetDeletionTimestamp() == nil {
		this.ctx.GetResourceCache().Set(RC_KEY_OPERATOR_POD, NewResourceCacheEntry(name, pod))
	} else {
		this.ctx.GetLog().WithValues("type", "Warning", "error", err).
			Info("Could not read operator's Pod resource. Will retry.")
	}
}

func (this *OperatorPodCF) Cleanup() bool {
	// No cleanup
	return true
}
