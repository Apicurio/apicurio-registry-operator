package impl

import (
	"strconv"

	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/services"
)

var _ loop.ControlLoop = &controlLoopImpl{}

type controlLoopImpl struct {
	ctx              context.LoopContext
	services         services.LoopServices
	controlFunctions []loop.ControlFunction
}

func NewControlLoopImpl(ctx context.LoopContext, services services.LoopServices) loop.ControlLoop {
	return &controlLoopImpl{
		ctx:              ctx,
		services:         services,
		controlFunctions: make([]loop.ControlFunction, 0, 32),
	}
}

func (this *controlLoopImpl) AddControlFunction(cf loop.ControlFunction) {
	this.controlFunctions = append(this.controlFunctions, cf)
}

func (this *controlLoopImpl) GetControlFunctions() []loop.ControlFunction {
	return this.controlFunctions
}

func (this *controlLoopImpl) Run() {
	this.services.BeforeRun()

	// CONTROL LOOP
	maxAttempts := len(this.GetControlFunctions()) * 2
	attempt := 0
	for ; attempt < maxAttempts; attempt++ {
		this.ctx.GetLog().Sugar().Infow("control loop executing",
			"attempt", strconv.Itoa(attempt), "maxAttempts", strconv.Itoa(maxAttempts))
		this.ctx.SetAttempts(attempt)
		// Run the CFs until we exceed the limit or the state has stabilized,
		// i.e. no action was taken by any CF
		stabilized := true
		for _, cf := range this.GetControlFunctions() {
			l := this.ctx.GetLog().Sugar().With("cf", cf.Describe())
			l.Debugw("control function sense")
			cf.Sense()
			l.Debugw("control function compare")
			discrepancy := cf.Compare()
			if discrepancy {
				l.Infow("control function respond")
				cf.Respond()
				stabilized = false
			}
		}

		if stabilized {
			this.ctx.GetLog().Info("Control loop is stable.")
			break
		}
		//this.ctx.GetLog().Info("Looping")
	}
	if attempt == maxAttempts {
		panic("Control loop stabilization limit exceeded.")
	}

	this.services.AfterRun()
}

func (this *controlLoopImpl) Cleanup() {
	// Perform resource cleanup

	this.ctx.GetLog().Sugar().Infow("ApicurioRegistry CR has been removed. Starting resource cleanup.",
		"app", this.ctx.GetAppName())
	maxAttempts := len(this.GetControlFunctions()) * 2
	attempt := 0
	for ; attempt < maxAttempts; attempt++ {
		finished := true
		for _, cf := range this.GetControlFunctions() {
			success := cf.Cleanup()
			if !success {
				this.ctx.GetLog().Sugar().Infow("Control function requested cleanup retry.",
					"cf", cf.Describe())
			}
			finished = finished && success
		}
		if finished {
			this.ctx.GetLog().Sugar().Infow("Cleanup finished successfully.",
				"app", this.ctx.GetAppName())
			break
		}
	}
	if attempt == maxAttempts {
		this.ctx.GetLog().Sugar().
			Warnw("Cleanup did not finish successfully. You may need to delete some of the resources manually.",
				"app", this.ctx.GetAppName())
	}
}

func (this *controlLoopImpl) GetContext() context.LoopContext {
	return this.ctx
}
