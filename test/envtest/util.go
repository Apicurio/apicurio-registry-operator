package envtest

import (
	"context"
	"go.uber.org/zap"
	core "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const T_SCALE = 4
const EVENTUALLY_CHECK_PERIOD = 500 * time.Millisecond

type SuiteState struct {
	log       *zap.Logger
	ctx       context.Context
	k8sClient client.Client
}

func isBefore(haystack []core.EnvVar, before string, after string) bool {
	beforeI := -1
	for i, _ := range haystack {
		if haystack[i].Name == before {
			beforeI = i
		}
	}
	afterI := -1
	for i, _ := range haystack {
		if haystack[i].Name == after {
			afterI = i
		}
	}
	return beforeI >= 0 && afterI >= 0 && afterI > beforeI
}

func get(haystack []core.EnvVar, name string) (*core.EnvVar, bool) {
	for i, _ := range haystack {
		if haystack[i].Name == name {
			return &haystack[i], true
		}
	}
	return nil, false
}
