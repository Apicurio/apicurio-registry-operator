package loop

import "github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"

type ControlLoop interface {
	AddControlFunction(cf ControlFunction)

	GetControlFunctions() []ControlFunction

	GetContext() *context.LoopContext

	Run()

	Cleanup()
}
