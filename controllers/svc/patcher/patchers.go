package patcher

import (
	"encoding/json"
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	c "github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/status"
	jsonpatch "github.com/evanphx/json-patch"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Patchers struct {
	kubePatcher *KubePatcher
	ocpPatcher  *OCPPatcher
}

func NewPatchers(ctx context.LoopContext, factoryKube *factory.KubeFactory, status *status.Status) *Patchers {
	this := &Patchers{}
	this.kubePatcher = NewKubePatcher(ctx, factoryKube, status)
	this.ocpPatcher = NewOCPPatcher(ctx)
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
	return jsonpatch.CreateMergePatch(o, n)
}

// Kind-of generic patching function to avoid repeating the code for each resource type
// TODO improve, there are better ways to patch objects, sigs_client.Client is preferrable to the rest client being used
func patchGeneric(
	ctx context.LoopContext,
	key string, // Resource cache key for the given resource
	genericToString func(interface{}) string, // Function to convert the resource to string (logging)
	genericType interface{}, // Empty instance of the resource struct
	typeString string, // A string representing the resource type (mostly, logging, see below)
	genericCreate func(meta.Object, c.Namespace, interface{}) (interface{}, error), // Function to create the resource using Kubernetes API
	genericPatch func(c.Namespace, c.Name, []byte) (interface{}, error), // Function to patch the resource using Kubernetes API
	genericGetName func(interface{}) c.Name) { // Function to get the resource name within k8s

	owner, exists := ctx.GetResourceCache().Get(resources.RC_KEY_SPEC)
	if !exists {
		ctx.GetLog().
			V(c.V_IMPORTANT).
			WithValues("resource", typeString).
			Info("Could not patch a resource. No ApicurioRegistry exists to set as the owner. Retrying.")
		ctx.SetRequeueNow()
		return
	}

	if entry, exists := ctx.GetResourceCache().Get(key); exists {

		namespace := ctx.GetAppNamespace()
		name := entry.GetName()
		value := entry.GetValue()
		// original := entry.GetOriginalValue() TODO

		// if exists
		if name != resources.RC_NOT_CREATED_NAME_EMPTY {
			// Skip actually if there are no PFs
			if !entry.HasChanged() {
				return
			}

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
				ctx.SetRequeueNow()
				return
			}

			// Optimization: Check if the patch is empty
			var patchJson map[string]interface{}
			if err := json.Unmarshal(patchData, &patchJson); err != nil {
				panic(err) // TODO
			}
			if patchJson == nil || len(patchJson) == 0 {
				//ctx.GetLog().WithValues("resource", typeString, "name", name, "patch", string(patchData)).
				//	Info("skipping empty patch")
				entry.ResetHasChanged()
				return
			}

			ctx.GetLog().WithValues("resource", typeString, "name", name).Info("patching")
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
				ctx.SetRequeueNow()
				return
			}
			// Reset PF after patching
			ctx.GetResourceCache().Set(key, resources.NewResourceCacheEntry(genericGetName(patched), patched))
		} else {
			ctx.GetLog().WithValues("resource", typeString).Info("Creating.")
			// Create it
			created, err := genericCreate(owner.GetValue().(*ar.ApicurioRegistry), namespace, value)
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
