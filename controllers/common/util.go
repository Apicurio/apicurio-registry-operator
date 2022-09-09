package common

import (
	"errors"
	"github.com/go-logr/logr"
	zaplogfmt "github.com/sykesm/zap-logfmt"
	uzap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"
	"time"
)

const (
	V_IMPORTANT = 0
	V_NORMAL    = 1
	V_DEBUG     = 2
	V_TRACE     = 3
)

func Fatal(log logr.Logger, err error, msg string) {
	log.Error(err, msg)
	panic("Fatal error, the operator can't recover.")
}

func FindString(haystack []string, needle string) (int, bool) {
	for i, v := range haystack {
		if needle == v {
			return i, true
		}
	}
	return -1, false
}

// TODO unnecessary
func FindStringKey(haystack map[string]bool, needle string) bool {
	for k, _ := range haystack {
		if needle == k {
			return true
		}
	}
	return false
}

func BuildLogger(devMode bool) logr.Logger {
	configLog := uzap.NewProductionEncoderConfig()
	configLog.EncodeTime = func(ts time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(ts.UTC().Format(time.RFC3339Nano))
	}
	logfmtEncoder := zaplogfmt.NewEncoder(configLog)
	logger := zap.New(zap.UseDevMode(devMode), zap.WriteTo(os.Stdout), zap.Encoder(logfmtEncoder))
	return logger
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
