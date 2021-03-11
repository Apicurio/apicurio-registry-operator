package impl

import (
	"strconv"

	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/services"
)

var _ loop.ControlLoop = &controlLoopImpl{}

type controlLoopImpl struct {
	ctx              *context.LoopContext
	services         *services.LoopServices
	controlFunctions []loop.ControlFunction
}

func NewControlLoopImpl(ctx *context.LoopContext, services *services.LoopServices) loop.ControlLoop {
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
		this.ctx.GetLog().WithValues("attempt", strconv.Itoa(attempt), "maxAttempts", strconv.Itoa(maxAttempts)).
			Info("Control loop executing.")
		// Run the CFs until we exceed the limit or the state has stabilized,
		// i.e. no action was taken by any CF
		stabilized := true
		for _, cf := range this.GetControlFunctions() {
			cf.Sense()
			discrepancy := cf.Compare()
			if discrepancy {
				this.ctx.GetLog().WithValues("cf", cf.Describe()).Info("Control function responding.")
				cf.Respond()
				stabilized = false
				break // Loop is restarted as soon as an action was taken
			}
		}

		if stabilized {
			this.ctx.GetLog().Info("Control loop is stable.")
			break
		}
	}
	if attempt == maxAttempts {
		panic("Control loop stabilization limit exceeded.")
	}

	this.services.AfterRun()
}

func (this *controlLoopImpl) Cleanup() {
	// Perform resource cleanup

	this.ctx.GetLog().WithValues("app", this.ctx.GetAppName()).Info("ApicurioRegistry CR has been removed. Starting resource cleanup.")
	maxAttempts := len(this.GetControlFunctions()) * 2
	attempt := 0
	for ; attempt < maxAttempts; attempt++ {
		finished := true
		for _, cf := range this.GetControlFunctions() {
			success := cf.Cleanup()
			if !success {
				this.ctx.GetLog().WithValues("cf", cf.Describe()).Info("Control function requested cleanup retry.")
			}
			finished = finished && success
		}
		if finished {
			this.ctx.GetLog().WithValues("app", this.ctx.GetAppName()).Info("Cleanup finished successfully.")
			break
		}
	}
	if attempt == maxAttempts {
		this.ctx.GetLog().WithValues("app", this.ctx.GetAppName(), "type", "Warning").
			Info("WARNING: Cleanup did not finish successfully. You may need to delete some of the resources manually.")
	}
}

func (this *controlLoopImpl) GetContext() *context.LoopContext {
	return this.ctx
}
