package resources

import "github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/common"

const RC_KEY_SPEC = "SPEC"
const RC_KEY_DEPLOYMENT = "DEPLOYMENT"
const RC_KEY_DEPLOYMENT_OCP = "DEPLOYMENT_OCP"
const RC_KEY_SERVICE = "SERVICE"
const RC_KEY_INGRESS = "INGRESS"
const RC_KEY_ROUTE_OCP = "ROUTE_OCP"
const RC_KEY_POD_DISRUPTION_BUDGET = "POD_DISRUPTION_BUDGET"
const RC_KEY_OPERATOR_POD = "OPERATOR_POD"

const RC_EMPTY_NAME = ""

// ===

// Return a MODIFIED COPY of the given resource. Do not modify or return the original value.
type PatchFunction = func(value interface{}) interface{}

// ===

type ResourceCacheEntry interface {
	// Get the k8s name of the resource, we are not supporting multiple namespaces yet
	// May return null if the resource was just created
	GetName() common.Name

	// get the stored value, it has to be cast using type assertion
	GetValue() interface{}

	GetOriginalValue() interface{}

	ApplyPatch(pf PatchFunction)

	IsPatched() bool
	ResetPatched()
	// TODO:
	// IsSingleValue() bool
	// GetValues() []ResourceCacheEntry
}

type resourceCacheEntry struct {
	name          common.Name
	value         interface{}
	originalValue interface{}
	isPatched     bool
}

func NewResourceCacheEntry(name common.Name, value interface{}) ResourceCacheEntry {
	this := &resourceCacheEntry{}
	this.name = name
	this.value = value
	this.originalValue = value
	this.isPatched = false
	return this
}

func (this *resourceCacheEntry) GetName() common.Name {
	return this.name
}

func (this *resourceCacheEntry) GetValue() interface{} {
	return this.value
}

func (this *resourceCacheEntry) GetOriginalValue() interface{} {
	return this.originalValue
}

func (this *resourceCacheEntry) ApplyPatch(pf PatchFunction) {
	this.value = pf(this.value)
	this.isPatched = true
}

func (this *resourceCacheEntry) IsPatched() bool {
	return this.isPatched
}

func (this *resourceCacheEntry) ResetPatched() {
	this.isPatched = false
}

// ===

type ResourceCache interface {
	// Returns the value and a boolean representing if the value exists
	Get(key string) (ResourceCacheEntry, bool)

	Set(key string, value ResourceCacheEntry) // TODO We may need multi-map in the future (add `type`)

	Remove(key string)

	Clear()
}

type resourceCache struct {
	cache map[string]ResourceCacheEntry
}

// Resource cache is a way for CF to avoid getting & recreating the resources unnecessarily,
// and do a pseudo-atomic patching
// It is cleared at the end of reconcilliation cycle, and there must be some CF responsible for filling the values
func NewResourceCache() ResourceCache {
	this := &resourceCache{}
	this.cache = make(map[string]ResourceCacheEntry)
	return this
}

func (this *resourceCache) Get(key string) (ResourceCacheEntry, bool) {
	value, exists := this.cache[key]
	return value, exists
}

func (this *resourceCache) Set(key string, value ResourceCacheEntry) {
	this.cache[key] = value
}

func (this *resourceCache) Remove(key string) {
	delete(this.cache, key)
}

func (this *resourceCache) Clear() {
	this.cache = make(map[string]ResourceCacheEntry)
}
