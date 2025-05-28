package common

import (
	"errors"
	zaplogfmt "github.com/sykesm/zap-logfmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	"os"
	"reflect"
	kzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"strings"
	"testing"
	"time"
)

var rootLog *zap.Logger = nil

func FindString(haystack []string, needle string) (int, bool) {
	for i, v := range haystack {
		if needle == v {
			return i, true
		}
	}
	return -1, false
}

func GetRootLogger(devMode bool) *zap.Logger {
	if rootLog != nil {
		return rootLog
	}
	level := os.Getenv("LOG_LEVEL")
	var zapLevel zapcore.Level
	switch strings.ToLower(level) {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn", "warning":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		if len(level) > 0 {
			_, _ = os.Stdout.WriteString("Invalid log level value '" + level + "'\n")
		}
		if devMode {
			zapLevel = zapcore.DebugLevel
		} else {
			zapLevel = zapcore.InfoLevel
		}
		_, _ = os.Stdout.WriteString("Setting log level to a default value '" + zapLevel.String() + "'\n")

	}
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(ts time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(ts.UTC().Format(time.RFC3339Nano))
	}
	encoder := zaplogfmt.NewEncoder(encoderConfig)
	rootLog = kzap.NewRaw(
		kzap.UseDevMode(devMode),
		kzap.WriteTo(os.Stdout),
		kzap.Encoder(encoder),
		kzap.Level(zapLevel))
	return rootLog
}

func AssertEquals(t *testing.T, expected interface{}, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Assertion failed, values ar not equal. Expected %v but got %v.", expected, actual)
	}
}

func FindIndex(haystack []interface{}, needle interface{}) (int, error) {
	for i, v := range haystack {
		if reflect.DeepEqual(v, needle) {
			return i, nil
		}
	}
	return -1, errors.New("Not found")
}

func IsInOrder(list []interface{}, before interface{}, after interface{}) bool {
	i1, e1 := FindIndex(list, before)
	i2, e2 := FindIndex(list, after)
	return e1 == nil && e2 == nil && i1 < i2
}

func AssertIsInOrder(t *testing.T, list []interface{}, before interface{}, after interface{}) {
	if !IsInOrder(list, before, after) {
		t.Errorf("Assertion failed, values are not in order.")
	}
}

func AssertSliceContains(t *testing.T, haystack []interface{}, needle interface{}) {
	if _, e := FindIndex(haystack, needle); e != nil {
		t.Errorf("Assertion failed, slice does not contain the value")
	}
}

// TODO Refactor
func SetVolumeInDeployment(deployment *apps.Deployment, volume *core.Volume) {
	SetVolume(&deployment.Spec.Template.Spec.Volumes, volume)
}

func SetVolume(volumes *[]core.Volume, volume *core.Volume) {
	volumeAlreadyExists := false

	// Modify volume if it exists, otherwise append as a new volume
	for i, vol := range *volumes {
		if vol.Name == volume.Name {
			volumeAlreadyExists = true
			(*volumes)[i] = *volume
			break
		}
	}
	if !volumeAlreadyExists {
		*volumes = append(*volumes, *volume)
	}
}

func SetVolumeMount(volumeMounts *[]core.VolumeMount, volumeMount *core.VolumeMount) {
	alreadyExists := false
	// Modify if it exists, otherwise append
	for i, v := range *volumeMounts {
		if v.Name == volumeMount.Name {
			alreadyExists = true
			(*volumeMounts)[i] = *volumeMount
			break
		}
	}
	if !alreadyExists {
		*volumeMounts = append(*volumeMounts, *volumeMount)
	}
}

func RemoveVolumeFromDeployment(deployment *apps.Deployment, volume *core.Volume) {

	deploymentVolumes := &deployment.Spec.Template.Spec.Volumes

	// Remove the Volume from the Deployment if it exists
	for i, vol := range *deploymentVolumes {
		if vol.Name == volume.Name {
			*deploymentVolumes = append((*deploymentVolumes)[:i], (*deploymentVolumes)[i+1:]...)
			break
		}
	}
}

func AddVolumeMountToContainer(container *core.Container, volumeMount *core.VolumeMount) {

	volumeMounts := &container.VolumeMounts
	mountAlreadyExists := false

	// Modify volumeMount if it exists, otherwise append as a new volumeMount
	for i, mount := range *volumeMounts {
		if mount.Name == volumeMount.Name {
			mountAlreadyExists = true
			(*volumeMounts)[i] = *volumeMount
			break
		}
	}
	if !mountAlreadyExists {
		*volumeMounts = append(*volumeMounts, *volumeMount)
	}
}

func RemoveVolumeMountFromContainer(container *core.Container, volumeMount *core.VolumeMount) {

	volumeMounts := &container.VolumeMounts

	// Remove the VolumeMount if it exists
	for i, mount := range *volumeMounts {
		if mount.MountPath == volumeMount.MountPath {
			*volumeMounts = append((*volumeMounts)[:i], (*volumeMounts)[i+1:]...)
			break
		}
	}
}

func AddContainerPort(ports *[]core.ContainerPort, port *core.ContainerPort) {
	exists := false
	for i, p := range *ports {
		if p.ContainerPort == port.ContainerPort {
			exists = true
			(*ports)[i] = *port
			break
		}
	}
	if !exists {
		*ports = append(*ports, *port)
	}
}

func RemovePortFromContainer(container *core.Container, port *core.ContainerPort) {
	ports := &container.Ports
	for i, p := range *ports {
		if p.ContainerPort == port.ContainerPort {
			*ports = append((*ports)[:i], (*ports)[i+1:]...)
			break
		}
	}
}

func AddPortToService(service *core.Service, port *core.ServicePort) {

	ports := &service.Spec.Ports
	portAlreadyExists := false

	// Modify servicePort if it exists, otherwise append as a new servicePort
	// TODO Maybe compare by name?
	for i, p := range *ports {
		if p.Port == port.Port {
			portAlreadyExists = true
			(*ports)[i] = *port
			break
		}
	}
	if !portAlreadyExists {
		*ports = append(*ports, *port)
	}
}

func RemovePortFromService(service *core.Service, port *core.ServicePort) {

	ports := &service.Spec.Ports

	// Remove servicePort if it exists
	for i, p := range *ports {
		if p.Port == port.Port {
			*ports = append((*ports)[:i], (*ports)[i+1:]...)
			break
		}
	}
}

func AddRuleToNetworkPolicy(policy *networking.NetworkPolicy, rule *networking.NetworkPolicyIngressRule) {
	exists := false
	for i, r := range policy.Spec.Ingress {
		equals := false
		if len(r.Ports) == len(rule.Ports) {
			for ii, p := range r.Ports {
				if p.Protocol != rule.Ports[ii].Protocol ||
					p.Port.String() != rule.Ports[ii].Port.String() {
					equals = false
					break
				}
			}
		}
		if equals {
			exists = true
			policy.Spec.Ingress[i] = *rule
			break
		}
	}
	if !exists {
		policy.Spec.Ingress = append(policy.Spec.Ingress, *rule)
	}
}

func RemoveRuleFromNetworkPolicy(policy *networking.NetworkPolicy, rule *networking.NetworkPolicyIngressRule) {
	for i, r := range policy.Spec.Ingress {
		equals := false
		if len(r.Ports) == len(rule.Ports) {
			for ii, p := range r.Ports {
				if p.Protocol != rule.Ports[ii].Protocol ||
					p.Port.String() != rule.Ports[ii].Port.String() {
					equals = true
					break
				}
			}
		}
		if equals {
			policy.Spec.Ingress = append(policy.Spec.Ingress[:i], policy.Spec.Ingress[i+1:]...)
			break
		}
	}
}

func SecretHasField(secret *core.Secret, field string) bool {

	for key, _ := range secret.Data {
		if key == field {
			return true
		}
	}
	return false
}

func HasPort(port string, ports []core.ServicePort) bool {

	for _, p := range ports {
		if p.Name == port {
			return true
		}
	}
	return false
}

// Return *true* if, for given source labels,
// the target label values exist and have the same value
func LabelsEqual(target map[string]string, source map[string]string) bool {
	for sourceKey, sourceValue := range source {
		targetValue, targetExists := target[sourceKey]
		if !targetExists || sourceValue != targetValue {
			return false
		}
	}
	return true
}

func LabelsDeleteUpdate(target *map[string]string, deleted map[string]string, updated map[string]string) {
	if *target == nil {
		*target = make(map[string]string)
	}
	if deleted != nil {
		for deletedKey, _ := range deleted {
			if _, targetExists := (*target)[deletedKey]; targetExists {
				delete(*target, deletedKey)
			}
		}
	}
	if updated != nil {
		for updatedKey, updatedValue := range updated {
			targetValue, targetExists := (*target)[updatedKey]
			if !targetExists || updatedValue != targetValue {
				(*target)[updatedKey] = updatedValue
			}
		}
	}
}

func LabelsUpdate(target *map[string]string, updated map[string]string) {
	LabelsDeleteUpdate(target, nil, updated)
}

func LabelsDelete(target *map[string]string, deleted map[string]string) {
	LabelsDeleteUpdate(target, deleted, nil)
}

func Copy[K comparable, V any](source map[K]V) map[K]V {
	c := make(map[K]V)
	for k, v := range source {
		c[k] = v
	}
	return c
}

func GetContainerByName(containers []core.Container, name string) *core.Container {
	for i, c := range containers {
		if c.Name == name {
			return &containers[i]
		}
	}
	return nil
}
