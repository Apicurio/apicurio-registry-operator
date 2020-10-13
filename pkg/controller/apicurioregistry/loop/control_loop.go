package loop

type ControlLoop interface {
	AddControlFunction(cf ControlFunction)

	GetControlFunctions() []ControlFunction

	GetContext() ControlLoopContext

	Run()

	Cleanup()
}
