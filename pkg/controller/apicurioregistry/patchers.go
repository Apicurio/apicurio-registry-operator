package apicurioregistry

import (
	"encoding/json"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

type Patchers struct {
	ctx         *Context
	kubePatcher *KubePatcher
	ocpPatcher  *OCPPatcher
}

func NewPatchers(ctx *Context) *Patchers {
	this := &Patchers{
		ctx: ctx,
	}
	this.kubePatcher = NewKubePatcher(ctx)
	this.ocpPatcher = NewOCPPatcher(ctx)
	return this
}

// =====

func (this *Patchers) OCP() *OCPPatcher {
	return this.ocpPatcher
}

func (this *Patchers) Kube() *KubePatcher {
	return this.kubePatcher
}

func (this *Patchers) Execute() {
	this.kubePatcher.Execute()
	this.ocpPatcher.Execute()
}

// =====

func createPatch(old, new, datastruct interface{}) ([]byte, error) {
	o, err := json.Marshal(old)
	if err != nil {
		return nil, err
	}
	n, err := json.Marshal(new)
	if err != nil {
		return nil, err
	}
	return strategicpatch.CreateTwoWayMergePatch(o, n, datastruct)
}

