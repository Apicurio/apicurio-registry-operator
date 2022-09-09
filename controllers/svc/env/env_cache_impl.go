package env

import (
	c "github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/go-logr/logr"
	core "k8s.io/api/core/v1"
	"reflect"
	"strings"
)

type envCacheEntry struct {
	value        *core.EnvVar
	dependencies []string
	priority     Priority
}

var _ EnvCacheEntry = &envCacheEntry{}

func (this *envCacheEntry) GetName() string {
	return this.value.Name
}

// The value is NOT copied
func (this *envCacheEntry) GetValue() *core.EnvVar {
	return this.value
}

// The value is NOT copied
func (this *envCacheEntry) GetDependencies() []string {
	return this.dependencies
}

func (this *envCacheEntry) GetPriority() Priority {
	return this.priority
}

// ===

type envCacheEntryBuilder struct {
	entry *envCacheEntry
}

var _ EnvCacheEntryBuilder = &envCacheEntryBuilder{}

// Makes a deep copy of the inner value
func NewEnvCacheEntryBuilder(value *core.EnvVar) EnvCacheEntryBuilder {
	builder := &envCacheEntryBuilder{
		entry: &envCacheEntry{
			value:        value.DeepCopy(),
			dependencies: make([]string, 0),
			priority:     PRIORITY_OPERATOR,
		},
	}
	return builder
}

func NewSimpleEnvCacheEntryBuilder(name string, value string) EnvCacheEntryBuilder {
	return NewEnvCacheEntryBuilder(&core.EnvVar{
		Name:  name,
		Value: value,
	})
}

func (this *envCacheEntryBuilder) SetDependency(name string) EnvCacheEntryBuilder {
	if _, exists := c.FindString(this.entry.dependencies, name); !exists {
		this.entry.dependencies = append(this.entry.dependencies, name)
	}
	// TODO Error?
	return this
}

func (this *envCacheEntryBuilder) SetPriority(priority Priority) EnvCacheEntryBuilder {
	this.entry.priority = priority
	return this
}

func (this *envCacheEntryBuilder) Build() EnvCacheEntry {
	if strings.TrimSpace(this.entry.GetName()) == "" {
		panic("Environment variable can not be empty nor whitespace")
	}
	return this.entry
}

// ===

type envCache struct {
	cache   map[string]EnvCacheEntry
	sorted  []EnvCacheEntry
	deleted map[string]bool
	changed bool
	log     logr.Logger
}

var _ EnvCache = &envCache{}

// The cache tries to preserve the order of addition.
// When sorting, it only reorders so far that it satisfies dependency relation
func NewEnvCache(log logr.Logger) EnvCache {
	return &envCache{
		cache:   make(map[string]EnvCacheEntry, 0),
		sorted:  make([]EnvCacheEntry, 0),
		deleted: make(map[string]bool, 0),
		changed: false,
		log:     log,
	}
}

// Try to get an existing entry.
// If it does not exist, the second return value is false.
func (this *envCache) Get(key string) (EnvCacheEntry, bool) {
	if this.WasDeleted(key) {
		return nil, false
	}
	v, exists := this.cache[key]
	return v, exists
}

// This function does not overwrite (and mark the cache as changed) when an existing
// env. value is found (using deep equal)
// OR the new entry has a lower priority
func (this *envCache) Set(value EnvCacheEntry) {
	if oldValue, exists := this.cache[value.GetName()]; exists && !this.WasDeleted(value.GetName()) {
		changed := !reflect.DeepEqual(value, oldValue)
		if value.GetPriority().toInt() >= oldValue.GetPriority().toInt() {
			// This will also update the priority
			if changed {
				this.cache[value.GetName()] = value
				this.changed = true
			}
		}
	} else {
		this.cache[value.GetName()] = value
		this.changed = true
	}
}

func (this *envCache) Delete(value EnvCacheEntry) bool {
	return this.DeleteByName(value.GetName())
}

func (this *envCache) DeleteByName(name string) bool {
	if _, exists := this.cache[name]; exists {
		this.deleted[name] = true
		this.changed = true
		return true
	}
	return false
}

func (this *envCache) WasDeleted(name string) bool {
	_, exists := this.deleted[name]
	return exists
}

func (this *envCache) Clear() {
	this.changed = true
	this.cache = make(map[string]EnvCacheEntry, 0)
	this.deleted = make(map[string]bool, 0)
}

func (this *envCache) ProcessAndAdvanceToNextPeriod() {
	this.changed = false
	for k, _ := range this.deleted {
		delete(this.cache, k)
	}
	this.deleted = make(map[string]bool, 0)
}

func (this *envCache) IsChanged() bool {
	return this.changed
}

// Copies the slice but not the values
func (this *envCache) GetSorted() []core.EnvVar {
	this.computeSorted()
	res := make([]core.EnvVar, 0)
	for _, v := range this.sorted {
		res = append(res, *v.GetValue())
	}
	return res
}

func (this *envCache) computeSorted() {
	this.sorted = make([]EnvCacheEntry, 0) // TODO Either use the cached value or use as a recursion argument
	processed := make(map[string]bool, len(this.cache))
	for p := 0; p <= PRIORITY_MAX.toInt(); p++ { // Make sure the variables are ordered by priority if possible
		for _, v := range this.cache {
			if v.GetPriority().toInt() == p {
				this.processWithDependencies(0, processed, v)
			}
		}
	}
}

func (this *envCache) processWithDependencies(depth int, processed map[string]bool, val EnvCacheEntry) {
	if depth > len(this.cache) {
		panic("Cycle detected during the processing of environment variables.")
	}
	if _, exists := processed[val.GetName()]; !exists && !this.WasDeleted(val.GetName()) { // Was not yet processed
		for _, dependencyName := range val.GetDependencies() { // Process & add deps first
			if d, exists := this.Get(dependencyName); exists {
				this.processWithDependencies(depth+1, processed, d)
			} else {
				this.log.Info("Dependency for an entry not found", "entryName", val.GetName(), "dependencyName", dependencyName)
			}
		}
		this.sorted = append(this.sorted, val)
		processed[val.GetName()] = true
	}
}
