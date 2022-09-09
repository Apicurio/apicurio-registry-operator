package env

import (
	core "k8s.io/api/core/v1"
)

type Priority int

const (
	PRIORITY_DEPLOYMENT Priority = 0
	PRIORITY_SPEC       Priority = 1
	PRIORITY_OPERATOR   Priority = 2
	PRIORITY_MAX        Priority = PRIORITY_OPERATOR
)

func (this Priority) toInt() int {
	return int(this)
}

type EnvCacheEntry interface {
	GetName() string

	GetValue() *core.EnvVar

	GetDependencies() []string

	GetPriority() Priority
}

type EnvCacheEntryBuilder interface {
	SetDependency(name string) EnvCacheEntryBuilder

	SetPriority(priority Priority) EnvCacheEntryBuilder

	Build() EnvCacheEntry
}

type EnvCache interface {
	// Returns the value and a boolean representing if the value exists
	Get(key string) (EnvCacheEntry, bool)

	Set(value EnvCacheEntry)

	// Get the entries based on the declared dependencies,
	// entry that has a dependency goes after it
	// panics if there is a dependency chain longer than 20 items
	GetSorted() []core.EnvVar

	// Try to delete and return true if the key existed
	Delete(value EnvCacheEntry) bool

	DeleteByName(key string) bool

	// Return true if the enty with the given key was marked fo deletion in the
	// given period
	WasDeleted(key string) bool

	// Delete the entries marked for deletion,
	// and mark the cache as not-changed again.
	// Do not call this outside of env CF!
	ProcessAndAdvanceToNextPeriod()

	IsChanged() bool
}
