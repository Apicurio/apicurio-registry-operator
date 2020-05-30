package apicurioregistry

// Defines an interface for a component responsible for some part of the
// reconciliation
type ControlFunction interface {
	// Get information from the environment
	Sense()

	// Compare the measured data with the intended state.
	// Do not affect the environment here, if possible.
	// If *true* there was a discrepancy and the next stage will be executed, else skipped
	Compare() bool

	// Do an action to get the system into alignment with the desired state
	Respond()

	// Return the description of the CF
	Describe() string
}
