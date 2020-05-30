package apicurioregistry

import (
	core "k8s.io/api/core/v1"
)

type EnvCacheEntry interface {
	GetName() string

	GetValue() *core.EnvVar

	SetInterpolationDependency(name string)

	GetDependencies() []string

	IsManaged() bool
}

type envCacheEntry struct {
	value        *core.EnvVar
	dependencies []string
	managed bool
}

var _ EnvCacheEntry = &envCacheEntry{}

func NewEnvCacheEntry(value *core.EnvVar) EnvCacheEntry {
	this := &envCacheEntry{}
	this.value = value
	this.dependencies = make([]string, 0) // TODO this may probably be a set
	this.managed = true
	return this
}

func NewEnvCacheEntryUnmanaged(value *core.EnvVar) EnvCacheEntry {
	this := &envCacheEntry{}
	this.value = value
	this.dependencies = make([]string, 0) // TODO this may probably be a set
	this.managed = false
	return this
}

func NewSimpleEnvCacheEntry(name string, value string) EnvCacheEntry {
	if name == "" {
		panic("Illegal argument.")
	}
	e := &core.EnvVar{
		Name:      name,
		Value:     value,
		ValueFrom: nil,
	}
	return NewEnvCacheEntry(e)
}

func (this *envCacheEntry) IsManaged() bool {
	return this.managed
}

func (this *envCacheEntry) GetName() string {
	return this.value.Name
}

func (this *envCacheEntry) GetValue() *core.EnvVar {
	return this.value
}

func (this *envCacheEntry) SetInterpolationDependency(name string) {
	if _, found := findString(this.dependencies, name); !found {
		this.dependencies = append(this.dependencies, name)
	}
}

func (this *envCacheEntry) GetDependencies() []string {
	return this.dependencies
}

// ===

type EnvCache interface {
	// Returns the value and a boolean representing if the value exists
	Get(key string) (EnvCacheEntry, bool)

	Set(value EnvCacheEntry)

	// Get the entries based on the declared dependencies,
	// entry that has a dependency goes after it
	// panics if there is a dependency chain longer than 20 items
	GetSorted() []core.EnvVar

	// Not used because the cache must keep the values between c. loops
	// Also resets changed mark and marks not ready
	// Clear()

	// Try to delete and return true if the key existed
	Delete(value EnvCacheEntry) bool

	// Do not call this outside of env CF!
	ResetChanged()

	IsChanged() bool
}

type envCache struct {
	cache   []EnvCacheEntry
	sorted  []EnvCacheEntry // thread safety
	changed bool
}

var _ EnvCache = &envCache{}

// The cache tries to preserve the order of addition.
// When sorting, it only reorders so far that it satisfies dependency relation
func NewEnvCache() EnvCache {
	this := &envCache{}
	this.cache = make([]EnvCacheEntry, 0)
	this.sorted = make([]EnvCacheEntry, 0)
	this.changed = false
	return this
}

func (this *envCache) find(key string) (int, EnvCacheEntry) {
	for i, v := range this.cache {
		if key == v.GetName() {
			return i, v
		}
	}
	return -1, nil
}

func (this *envCache) Get(key string) (EnvCacheEntry, bool) {
	i, v := this.find(key)
	return v, i >= 0
}

func (this *envCache) Set(value EnvCacheEntry) {
	this.changed = true // We do not need to inspect contents
	if value.GetName() == "" {
		panic("Illegal argument: Cannot set env. variable with an empty name")
	}
	if i, _ := this.find(value.GetName()); i >= 0 {
		this.cache[i] = value
	} else {
		this.cache = append(this.cache, value)
	}
}

func (this *envCache) Delete(value EnvCacheEntry) bool {
	if i, _ := this.find(value.GetName()); i >= 0 {
		if i < len(this.cache)-1 {
			this.cache = append(this.cache[:i], this.cache[i+1:]...)
		}
		this.cache = this.cache[:len(this.cache)-1]
		return true
	}
	return false
}

func (this *envCache) Clear() {
	this.changed = false
	this.cache = make([]EnvCacheEntry, 0)
}

func (this *envCache) ResetChanged() {
	this.changed = false
}

func (this *envCache) IsChanged() bool {
	return this.changed
}

func (this *envCache) getSorted() []EnvCacheEntry {
	this.sorted = make([]EnvCacheEntry, 0)
	processed := make(map[string]bool, len(this.cache))
	for _, v := range this.cache {
		this.processWithDependencies(0, processed, v)
	}
	return this.sorted
}

func (this *envCache) GetSorted() []core.EnvVar {
	sorted := this.getSorted()
	res := make([]core.EnvVar, 0)
	for _, v := range sorted {
		res = append(res, *v.GetValue())
	}
	return res
}

func (this *envCache) processWithDependencies(depth int, proc map[string]bool, val EnvCacheEntry) {
	if depth > 20 {
		panic("Recursion is too deep or there is a cycle.")
	}
	if !findStringKey(proc, val.GetName()) { // Was not yet processed
		for _, name := range val.GetDependencies() { // Process & add deps first
			d, _ := this.Get(name)
			this.processWithDependencies(depth+1, proc, d)
		}
		this.sorted = append(this.sorted, val)
		proc[val.GetName()] = true
	}
}
