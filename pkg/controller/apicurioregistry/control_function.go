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

	// Release & Cleanup the resources created by this CF.
	// Return *true* if the cleanup was successful or is not needed.
	// If a CF returns false, the cleanup will be reattempted several times,
	// mostly in case other CFs have to do their cleanup first.
	// Warning: This function may be executed multiple times even if it returned *true*.
	Cleanup() bool
}
