package loop

import "github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"

type ControlLoop interface {
	AddControlFunction(cf ControlFunction)

	GetControlFunctions() []ControlFunction

	GetContext() context.LoopContext

	Run()

	Cleanup()
}
