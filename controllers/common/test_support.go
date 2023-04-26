package common

import (
	"go.uber.org/zap"
	"time"
)

type TestSupport struct {
	enabled                     bool
	canMakeHTTPRequestToOperand bool
	operandMetricsReportReady   bool
	loopTick                    time.Time
	log                         *zap.Logger
	features                    *SupportedFeatures
}

func NewTestSupport(rootLog *zap.Logger, enabled bool) *TestSupport {
	log := rootLog.Named("testing")
	if enabled {
		log.Sugar().Warnw("TESTING SUPPORT IS ENABLED. YOU SHOULD NOT SEE THIS MESSAGE IN PRODUCTION.")
	}
	return &TestSupport{
		enabled:                     enabled,
		canMakeHTTPRequestToOperand: false,
		operandMetricsReportReady:   false,
		loopTick:                    time.Time{},
		log:                         log,
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

func (this *TestSupport) SetMockCanMakeHTTPRequestToOperand(value bool) {
	this.panicIfNotTesting()
	this.canMakeHTTPRequestToOperand = value
}

func (this *TestSupport) GetMockCanMakeHTTPRequestToOperand() bool {
	this.panicIfNotTesting()
	return this.canMakeHTTPRequestToOperand
}

func (this *TestSupport) SetMockOperandMetricsReportReady(value bool) {
	this.panicIfNotTesting()
	this.operandMetricsReportReady = value
}

func (this *TestSupport) GetMockOperandMetricsReportReady() bool {
	this.panicIfNotTesting()
	return this.operandMetricsReportReady
}

func (this *TestSupport) ResetTimer() {
	this.panicIfNotTesting()
	this.loopTick = time.Now()
}

func (this *TestSupport) TimerDuration() time.Duration {
	this.panicIfNotTesting()
	return time.Now().Sub(this.loopTick)
}

func (this *TestSupport) SetSupportedFeatures(features *SupportedFeatures) {
	this.features = features
}

func (this *TestSupport) GetSupportedFeatures() *SupportedFeatures {
	return this.features
}
