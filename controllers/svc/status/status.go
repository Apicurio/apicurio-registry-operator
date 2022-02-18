package status

import (
	"strconv"

	api "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/status/conditions"
)

// status
const CFG_STA_IMAGE = "CFG_STA_IMAGE"
const CFG_STA_DEPLOYMENT_NAME = "CFG_STA_DEPLOYMENT_NAME"
const CFG_STA_SERVICE_NAME = "CFG_STA_SERVICE_NAME"
const CFG_STA_INGRESS_NAME = "CFG_STA_INGRESS_NAME"
const CFG_STA_NETWORK_POLICY_NAME = "CFG_STA_NETWORK_POLICY_NAME"
const CFG_STA_REPLICA_COUNT = "CFG_STA_REPLICA_COUNT"
const CFG_STA_ROUTE = "CFG_STA_ROUTE"

type Status struct {
	config     map[string]string
	ctx        *context.LoopContext
	conditions conditions.ConditionManager
}

func NewStatus(ctx *context.LoopContext, conditions conditions.ConditionManager) *Status {

	this := &Status{
		config:     make(map[string]string),
		ctx:        ctx,
		conditions: conditions,
	}
	this.init()
	return this
}

func (this *Status) init() {
	// DO NOT USE `spec` ! It's nil at this point
	// status
	this.set(this.config, CFG_STA_IMAGE, "")
	this.set(this.config, CFG_STA_DEPLOYMENT_NAME, "")
	this.set(this.config, CFG_STA_SERVICE_NAME, "")
	this.set(this.config, CFG_STA_INGRESS_NAME, "")
	this.set(this.config, CFG_STA_REPLICA_COUNT, "")
	this.set(this.config, CFG_STA_ROUTE, "")
}

// =====

func (this *Status) set(mapp map[string]string, key string, value string) {
	ptr := &value
	if key == "" {
		panic("Fatal: Empty key for " + *ptr)
	}
	mapp[key] = *ptr
}

func (this *Status) setDefault(mapp map[string]string, key string, value string, defaultValue string) {
	if value == "" {
		value = defaultValue
	}
	this.set(mapp, key, value)
}

func (this *Status) SetConfig(key string, value string) {
	this.set(this.config, key, value)
}

func (this *Status) SetConfigInt32P(key string, value *int32) {
	this.set(this.config, key, strconv.FormatInt(int64(*value), 10))
}

func (this *Status) GetConfig(key string) string {
	v, ok := this.config[key]
	if !ok {
		panic("Fatal: Status key '" + key + "' not found.")
	}
	return v
}

func (this *Status) GetConfigInt32P(key string) *int32 {
	i, _ := strconv.ParseInt(this.GetConfig(key), 10, 32)
	i2 := int32(i)
	return &i2
}

func (this *Status) ComputeStatus() {
	entry, exists := this.ctx.GetResourceCache().Get(resources.RC_KEY_STATUS)
	if exists {
		entry.ApplyPatch(func(value interface{}) interface{} {
			status := value.(*api.ApicurioRegistryStatus)

			// Info
			status.Info.Host = this.GetConfig(CFG_STA_ROUTE)

			// Conditions
			status.Conditions = this.conditions.Execute()

			// Resources
			// TODO Refactor
			res := make([]api.ApicurioRegistryStatusManagedResource, 0)
			if this.GetConfig(CFG_STA_DEPLOYMENT_NAME) != "" {
				res = append(res, api.ApicurioRegistryStatusManagedResource{
					Kind:      "Deployment",
					Namespace: this.ctx.GetAppNamespace().Str(),
					Name:      this.GetConfig(CFG_STA_DEPLOYMENT_NAME),
				})
			}
			if this.GetConfig(CFG_STA_SERVICE_NAME) != "" {
				res = append(res, api.ApicurioRegistryStatusManagedResource{
					Kind:      "Service",
					Namespace: this.ctx.GetAppNamespace().Str(),
					Name:      this.GetConfig(CFG_STA_SERVICE_NAME),
				})
			}
			if this.GetConfig(CFG_STA_INGRESS_NAME) != "" {
				res = append(res, api.ApicurioRegistryStatusManagedResource{
					Kind:      "Ingress",
					Namespace: this.ctx.GetAppNamespace().Str(),
					Name:      this.GetConfig(CFG_STA_INGRESS_NAME),
				})
			}
			status.ManagedResources = res

			return status
		})
	}
}
