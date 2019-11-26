package apicurioregistry

import (
	ar "github.com/apicurio/apicurio-operators/apicurio-registry/pkg/apis/apicur/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"strconv"
	"strings"
)

// env
const ENV_QUARKUS_PROFILE = "QUARKUS_PROFILE"

const ENV_QUARKUS_DATASOURCE_URL = "QUARKUS_DATASOURCE_URL"
const ENV_QUARKUS_DATASOURCE_USERNAME = "QUARKUS_DATASOURCE_USERNAME"
const ENV_QUARKUS_DATASOURCE_PASSWORD = "QUARKUS_DATASOURCE_PASSWORD"

const ENV_KAFKA_BOOTSTRAP_SERVERS = "KAFKA_BOOTSTRAP_SERVERS"

const ENV_APPLICATION_SERVER = "APPLICATION_SERVER"
const ENV_APPLICATION_ID = "APPLICATION_ID"

// cfg
const CFG_PERSISTENCE_TYPE = "PERSISTENCE_TYPE"
const CFG_IMAGE_REGISTRY = "IMAGE_REGISTRY"
const CFG_IMAGE_VERSION = "IMAGE_VERSION"

// dep
const CFG_DEP_REPLICAS = "REPLICAS"
const CFG_DEP_ROUTE = "ROUTE"
const CFG_DEP_CPU_REQUESTS = "CPU_REQUESTS"
const CFG_DEP_CPU_LIMIT = "CPU_LIMIT"
const CFG_DEP_MEMORY_REQUESTS = "MEMORY_REQUESTS"
const CFG_DEP_MEMORY_LIMIT = "MEMORY_LIMIT"

// status
const CFG_STA_IMAGE = "CFG_STA_IMAGE"
const CFG_STA_DEPLOYMENT_NAME = "CFG_STA_DEPLOYMENT_NAME"
const CFG_STA_SERVICE_NAME = "CFG_STA_SERVICE_NAME"
const CFG_STA_INGRESS_NAME = "CFG_STA_INGRESS_NAME"
const CFG_STA_REPLICA_COUNT = "CFG_STA_REPLICA_COUNT"

type Configuration struct {
	spec      *ar.ApicurioRegistry
	config    map[string]string
	envConfig map[string]string
	errors    *[]string
}

func NewConfiguration() *Configuration {

	res := &Configuration{
		//spec:      spec,
		config:    make(map[string]string),
		envConfig: make(map[string]string),
		errors:    new([]string),
	}
	res.init()
	//res.update()
	return res
}

func (this *Configuration) Update(spec *ar.ApicurioRegistry) {
	if spec == nil {
		panic("Run 'Update' after constructing a new instance.")
	}
	this.spec = spec
	this.update()
}

// this runs every loop!
func (this *Configuration) update() {
	this.set(this.envConfig, ENV_QUARKUS_PROFILE, "prod", required)

	this.set(this.config, CFG_IMAGE_REGISTRY, this.spec.Spec.Image.Registry, required)
	this.set(this.config, CFG_IMAGE_VERSION, this.spec.Spec.Image.Version, required)

	this.set(this.config, CFG_PERSISTENCE_TYPE, this.spec.Spec.Configuration.Persistence, enum("mem", "jpa", "kafka", "streams"))

	if "jpa" == this.spec.Spec.Configuration.Persistence {
		this.set(this.envConfig, ENV_QUARKUS_DATASOURCE_URL, this.spec.Spec.Configuration.DataSource.Url, required)
		this.set(this.envConfig, ENV_QUARKUS_DATASOURCE_USERNAME, this.spec.Spec.Configuration.DataSource.UserName, required)
		this.set(this.envConfig, ENV_QUARKUS_DATASOURCE_PASSWORD, this.spec.Spec.Configuration.DataSource.Password, required)
	}
	if "kafka" == this.spec.Spec.Configuration.Persistence {
		this.set(this.envConfig, ENV_KAFKA_BOOTSTRAP_SERVERS, this.spec.Spec.Configuration.Kafka.BootstrapServers, required)
	}
	if "streams" == this.spec.Spec.Configuration.Persistence {
		this.set(this.envConfig, ENV_KAFKA_BOOTSTRAP_SERVERS, this.spec.Spec.Configuration.Streams.BootstrapServers, required)
		this.set(this.envConfig, ENV_APPLICATION_SERVER, this.spec.Spec.Configuration.Streams.ApplicationServer, required)
		this.set(this.envConfig, ENV_APPLICATION_ID, this.spec.Spec.Configuration.Streams.ApplicationId, required)
	}

	if this.spec.Spec.Deployment.Replicas == 0 {
		this.spec.Spec.Deployment.Replicas = 2
	}
	this.set(this.config, CFG_DEP_REPLICAS, strconv.FormatInt(int64(this.spec.Spec.Deployment.Replicas), 10), required) // :(
	this.set(this.config, CFG_DEP_ROUTE, this.spec.Spec.Deployment.Route, defaultValue("registry.example.com"))

	this.set(this.config, CFG_DEP_CPU_REQUESTS, this.spec.Spec.Deployment.Resources.Cpu.Requests, defaultValue("0.1"))
	this.set(this.config, CFG_DEP_CPU_LIMIT, this.spec.Spec.Deployment.Resources.Cpu.Limit, defaultValue("1"))

	this.set(this.config, CFG_DEP_MEMORY_REQUESTS, this.spec.Spec.Deployment.Resources.Memory.Requests, defaultValue("600Mi"))
	this.set(this.config, CFG_DEP_MEMORY_LIMIT, this.spec.Spec.Deployment.Resources.Memory.Limit, defaultValue("1300Mi"))
}

func (this *Configuration) init() {
	// DO NOT USE 'spec' ! It's nil at this point
	// status
	this.set(this.config, CFG_STA_IMAGE, "", noOp)
	this.set(this.config, CFG_STA_DEPLOYMENT_NAME, "", noOp)
	this.set(this.config, CFG_STA_SERVICE_NAME, "", noOp)
	this.set(this.config, CFG_STA_INGRESS_NAME, "", noOp)
	this.set(this.config, CFG_STA_REPLICA_COUNT, "", noOp)
}

func (this *Configuration) fail(error string) {
	//foo := append(*this.errors, "Configuration error: "+error)
	//this.errors = &foo
	// TODO ^
	panic(error)
}

// TODO remove ?
func (this *Configuration) GetErrors() (errorsPresent bool, errors string) {
	if len(*this.errors) > 0 {
		errors := ""
		for _, v := range *this.errors {
			errors = errors + v + ". "
		}
		return true, errors
	}
	return false, ""
}

func (this *Configuration) set(mapp map[string]string, key string, value string, validate func(*string) (bool, string)) {
	ptr := &value
	if key == "" {
		panic("Empty key for " + *ptr)
	}
	ok, errStr := validate(ptr)
	if ok {
		//if *ptr == "" {
		//	panic("Empty value for key " + key)
		//}
		mapp[key] = *ptr
	} else {
		this.fail("Value '" + *ptr + "' for key '" + key + "' is not valid: " + errStr)
	}
}

// =====

func (this *Configuration) SetConfig(key string, value string) {
	this.set(this.config, key, value, required)
}

func (this *Configuration) SetConfigInt32P(key string, value *int32) {
	this.set(this.config, key, strconv.FormatInt(int64(*value), 10), required)
}

func (this *Configuration) GetConfig(key string) string {
	v, ok := this.config[key]
	if !ok {
		panic("Fatal: Configuration key '" + key + "' not found.")
		// TODO maybe just return "" ?
	}
	return v
}

func (this *Configuration) GetConfigInt32P(key string) *int32 {
	i, _ := strconv.ParseInt(this.GetConfig(key), 10, 32)
	i2 := int32(i)
	return &i2
}

// =====

func required(value *string) (bool, string) {
	return *value != "", "Value is empty."
}

func defaultValue(defaultStr string) (func(*string) (bool, string)) {
	return func(value *string) (bool, string) {
		if *value == "" {
			*value = defaultStr
		}
		return true, ""
	}
}

func noOp(value *string) (bool, string) {
	return true, ""
}

func enum(enums ...string) func(*string) (bool, string) {
	return func(value *string) (bool, string) {
		for _, e := range enums {
			if e == *value {
				return true, ""
			}
		}
		return false, "Value is not one of '" + strings.Join(enums, ", ") + "'."
	}
}

// =====

func (this *Configuration) GetImage() string {
	return this.spec.Spec.Image.Registry + "/" +
		"apicurio-registry-" + strings.ToLower(this.spec.Spec.Configuration.Persistence) + ":" +
		this.spec.Spec.Image.Version
}

func (this *Configuration) getEnv() []corev1.EnvVar {
	var env = []corev1.EnvVar{}
	for k, v := range this.envConfig {
		env = append(env, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	return env
}

func (this *Configuration) GetSpecName() string {
	return this.spec.Name
}

func (this *Configuration) GetSpecNamespace() string {
	return this.spec.Namespace
}
