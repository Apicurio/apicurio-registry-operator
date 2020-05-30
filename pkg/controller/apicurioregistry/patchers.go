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

func (this *Patchers) Reload() {
	this.kubePatcher.Reload()
	this.ocpPatcher.Reload()
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

// Kind-of generic patching function to avoid repeating the code for each resource type
func patchGeneric(
	ctx *Context,
	key string,
	genericGet func(string, string) (interface{}, error),
	genericToString func(interface{}) string,
	genericType interface{},
	typeString string,
	genericCreate func(string, interface{}) (interface{}, error),
	genericPatch func(string, string, []byte) (interface{}, error),
	genericGetName func(interface{}) string,
    removeStatus func(interface{}) interface{}) {

	if entry, exists := ctx.GetResourceCache().Get(key); exists {

		namespace := ctx.GetConfiguration().GetAppNamespace()
		name := entry.GetName()
		value := entry.GetValue()
		// original := entry.GetOriginalValue() TODO

		// if exists
		if name != RC_EMPTY_NAME {
			// Skip actually if there are no PFs
			if !entry.IsPatched() {
				return
			}

			ctx.GetLog().WithValues("resource", typeString, "name", name).Info("Patching.")

			// get the resource
			actualValue, err := genericGet(namespace, name)
			if err != nil {
				ctx.GetLog().
					WithValues("type", "Warning", "resource", typeString, "name", name, "error", err).
					Info("Could not get existing resource.")
				ctx.GetResourceCache().Remove(key) // Re-create the resource
				ctx.SetRequeue()
				return
			}
			patchData0, err := createPatch(removeStatus(actualValue), removeStatus(value), genericType)

			patchData := patchData0
			if(typeString != "ar.ApicurioRegistry") {
				patchData = append(make([]byte, 0), "{\"spec\":"...)
				patchData = append(patchData, patchData0...)
				patchData = append(patchData, "}"...)
			}

			if err != nil {
				ctx.GetLog().
					WithValues("type", "Warning", "resource", typeString, "error", err,
						"name", name, "original", genericToString(actualValue), "target", genericToString(value)).
					Info("Could not create patch data.")
				// Remove patch changes...
				// ctx.GetResourceCache().Set(key, NewResourceCacheEntry(genericGetName(original), original)) TODO
				ctx.GetResourceCache().Remove(key)
				ctx.SetRequeue()
				return
			}
			patched, err := genericPatch(namespace, name, patchData)
			if err != nil {
				// Could not apply patch. Maybe it was modified by external source.
				ctx.GetLog().
					WithValues("type", "Warning", "resource", typeString, "error", err,
						"name", name, "original", genericToString(actualValue), "target", genericToString(value),
						"patch", string(patchData)).
					Info("Could not submit patch.")
				// Remove patch changes
				// ctx.GetResourceCache().Set(key, NewResourceCacheEntry(genericGetName(original), original)) TODO
				ctx.GetResourceCache().Remove(key)
				ctx.SetRequeue()
				return
			}
			// Reset PF after patching
			ctx.GetResourceCache().Set(key, NewResourceCacheEntry(genericGetName(patched), patched))
		} else {
			ctx.GetLog().WithValues("resource", typeString).Info("Creating.")
			// Create it
			created, err := genericCreate(namespace, value)
			if err != nil {
				// Could not create.
				// Delete the value from cache so it can be tried again
				ctx.GetLog().
					WithValues("type", "Warning", "resource", typeString, "error", err,
						"target", genericToString(value)).
					Info("Could not create new resource.")
				ctx.GetResourceCache().Remove(key)
				return
			}
			// Reset PF
			ctx.GetResourceCache().Set(key, NewResourceCacheEntry(genericGetName(created), created))
		}
	}
}
