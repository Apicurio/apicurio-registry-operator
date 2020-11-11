package patcher

import (
	"encoding/json"

	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/common"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/client"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

type Patchers struct {
	kubePatcher *KubePatcher
	ocpPatcher  *OCPPatcher
}

func NewPatchers(ctx *context.LoopContext, clients *client.Clients, factoryKube *factory.KubeFactory) *Patchers {
	this := &Patchers{}
	this.kubePatcher = NewKubePatcher(ctx, clients, factoryKube)
	this.ocpPatcher = NewOCPPatcher(ctx, clients)
	return this
}

// =====

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
// TODO improve, there are better ways to patch objects, sigs_client.Client is preferrable to the rest client being used
func patchGeneric(
	ctx *context.LoopContext,
	key string, // Resource cache key for the given resource
	genericToString func(interface{}) string, // Function to convert the resource to string (logging)
	genericType interface{}, // Empty instance of the resource struct
	typeString string, // A string representing the resource type (mostly, logging, see below)
	genericCreate func(common.Namespace, interface{}) (interface{}, error), // Function to create the resource using Kubernetes API
	genericPatch func(common.Namespace, common.Name, []byte) (interface{}, error), // Function to patch the resource using Kubernetes API
	genericGetName func(interface{}) common.Name) { // Function to get the resource name within k8s

	if entry, exists := ctx.GetResourceCache().Get(key); exists {

		namespace := ctx.GetAppNamespace()
		name := entry.GetName()
		value := entry.GetValue()
		// original := entry.GetOriginalValue() TODO

		// if exists
		if name != resources.RC_EMPTY_NAME {
			// Skip actually if there are no PFs
			if !entry.IsPatched() {
				return
			}

			ctx.GetLog().WithValues("resource", typeString, "name", name).Info("Patching.")

			actualValue := entry.GetOriginalValue()
			patchData, err := createPatch(actualValue, value, genericType)

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
			ctx.GetResourceCache().Set(key, resources.NewResourceCacheEntry(genericGetName(patched), patched))
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
			ctx.GetResourceCache().Set(key, resources.NewResourceCacheEntry(genericGetName(created), created))
		}
	}
}
