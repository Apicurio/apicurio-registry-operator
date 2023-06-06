package common

import (
	"go.uber.org/zap"
	"time"
)

type TestSupport struct {
	enabled    bool
	log        *zap.Logger
	features   *SupportedFeatures
	namespaced map[string]*testSupportNamespaced
}

type testSupportNamespaced struct {
	canMakeHTTPRequestToOperand bool
	operandMetricsReportReady   bool
	loopTick                    time.Time
}

func NewTestSupport(rootLog *zap.Logger, enabled bool) *TestSupport {
	log := rootLog.Named("testing")
	if enabled {
		log.Sugar().Warnw("TESTING SUPPORT IS ENABLED. YOU SHOULD NOT SEE THIS MESSAGE IN PRODUCTION.")
	}
	return &TestSupport{
		enabled:    enabled,
		namespaced: make(map[string]*testSupportNamespaced),
		log:        log,
	}
}

func newTestSupportNamespaced() *testSupportNamespaced {
	return &testSupportNamespaced{
		canMakeHTTPRequestToOperand: false,
		operandMetricsReportReady:   false,
		loopTick:                    time.Time{},
	}
}

func (this *TestSupport) IsEnabled() bool {
	return this.enabled
}

func (this *TestSupport) panicIfNotTesting() {
	if !this.enabled {
		panic("TESTING SUPPORT IS ENABLED.")
	}
}

func (this *TestSupport) SetMockCanMakeHTTPRequestToOperand(namespace string, value bool) {
	this.panicIfNotTesting()
	if _, e := this.namespaced[namespace]; !e {
		this.namespaced[namespace] = newTestSupportNamespaced()
	}
	this.namespaced[namespace].canMakeHTTPRequestToOperand = value
}

func (this *TestSupport) GetMockCanMakeHTTPRequestToOperand(namespace string) bool {
	this.panicIfNotTesting()
	if _, e := this.namespaced[namespace]; !e {
		this.namespaced[namespace] = newTestSupportNamespaced()
	}
	return this.namespaced[namespace].canMakeHTTPRequestToOperand
}

func (this *TestSupport) SetMockOperandMetricsReportReady(namespace string, value bool) {
	this.panicIfNotTesting()
	if _, e := this.namespaced[namespace]; !e {
		this.namespaced[namespace] = newTestSupportNamespaced()
	}
	this.namespaced[namespace].operandMetricsReportReady = value
}

func (this *TestSupport) GetMockOperandMetricsReportReady(namespace string) bool {
	this.panicIfNotTesting()
	if _, e := this.namespaced[namespace]; !e {
		this.namespaced[namespace] = newTestSupportNamespaced()
	}
	return this.namespaced[namespace].operandMetricsReportReady
}

func (this *TestSupport) ResetTimer(namespace string) {
	this.panicIfNotTesting()
	if _, e := this.namespaced[namespace]; !e {
		this.namespaced[namespace] = newTestSupportNamespaced()
	}
	this.namespaced[namespace].loopTick = time.Now()
}

func (this *TestSupport) TimerDuration(namespace string) time.Duration {
	this.panicIfNotTesting()
	if _, e := this.namespaced[namespace]; !e {
		this.namespaced[namespace] = newTestSupportNamespaced()
	}
	return time.Now().Sub(this.namespaced[namespace].loopTick)
}

func (this *TestSupport) SetSupportedFeatures(features *SupportedFeatures) {
	this.features = features
}

func (this *TestSupport) GetSupportedFeatures() *SupportedFeatures {
	return this.features
}
