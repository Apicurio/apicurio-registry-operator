package common

import (
	"errors"
	zaplogfmt "github.com/sykesm/zap-logfmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
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

func First(first interface{}, _ interface{}) interface{} {
	return first
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

func AddVolumeToDeployment(deployment *apps.Deployment, volume *core.Volume) {

	deploymentVolumes := &deployment.Spec.Template.Spec.Volumes
	volumeAlreadyExists := false

	// Modify volume if it exists, otherwise append as a new volume
	for i, vol := range *deploymentVolumes {
		if vol.Name == volume.Name {
			volumeAlreadyExists = true
			(*deploymentVolumes)[i] = *volume
			break
		}
	}
	if !volumeAlreadyExists {
		*deploymentVolumes = append(*deploymentVolumes, *volume)
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

func AddEnvVarToContainer(container *core.Container, envVar *core.EnvVar) {

	envVars := &container.Env
	varAlreadyExists := false

	// Modify envVar if it exists, otherwise append as a new envVar
	for i, variable := range *envVars {
		if variable.Name == envVar.Name {
			varAlreadyExists = true
			(*envVars)[i] = *envVar
			break
		}
	}
	if !varAlreadyExists {
		*envVars = append(*envVars, *envVar)
	}

}

func RemoveEnvVarFromContainer(container *core.Container, envVar *core.EnvVar) {

	envVars := &container.Env

	// Remove envVar from container
	for i, variable := range *envVars {
		if variable.Name == envVar.Name {
			*envVars = append((*envVars)[:i], (*envVars)[i+1:]...)
			break
		}
	}

}

func AddPortToContainer(container *core.Container, port *core.ContainerPort) {

	ports := &container.Ports
	portAlreadyExists := false

	// Modify containerPort if it exists, otherwise append as a new containerPort
	for i, p := range *ports {
		if p.ContainerPort == port.ContainerPort {
			portAlreadyExists = true
			(*ports)[i] = *port
			break
		}
	}
	if !portAlreadyExists {
		*ports = append(*ports, *port)
	}

}

func AddPortToService(service *core.Service, port *core.ServicePort) {

	ports := &service.Spec.Ports
	portAlreadyExists := false

	// Modify servicePort if it exists, otherwise append as a new servicePort
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

func SecretHasTLSFields(secret *core.Secret) bool {

	foundCert := false
	foundKey := false

	// Validate the secret has both the 'tls.crt' and 'tls.key' fields
	for key, _ := range secret.Data {
		if key == "tls.crt" {
			foundCert = true
		} else if key == "tls.key" {
			foundKey = true
		}
	}
	if !foundCert || !foundKey {
		return false
	}
	return true

}

func HasPort(port string, ports []core.ServicePort) bool {

	for _, p := range ports {
		if p.Name == port {
			return true
		}
	}
	return false

}
