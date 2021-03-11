package status

import (
	"github.com/go-logr/logr"
	"strconv"
)

// status
const CFG_STA_IMAGE = "CFG_STA_IMAGE"
const CFG_STA_DEPLOYMENT_NAME = "CFG_STA_DEPLOYMENT_NAME"
const CFG_STA_SERVICE_NAME = "CFG_STA_SERVICE_NAME"
const CFG_STA_INGRESS_NAME = "CFG_STA_INGRESS_NAME"
const CFG_STA_REPLICA_COUNT = "CFG_STA_REPLICA_COUNT"
const CFG_STA_ROUTE = "CFG_STA_ROUTE"

type Status struct {
	config map[string]string
	log    logr.Logger
}

// This is at the moment only used for status vars.
// TODO Refactor
func NewStatus(log logr.Logger) *Status {

	res := &Status{
		config: make(map[string]string),
		log:    log,
	}
	res.init()
	return res
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
