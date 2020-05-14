package apicurioregistry

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type SenseRes struct {
	err *error
}

type CompareRes struct {
	discrepancy bool
	err         *error
}

type RespondRes struct {
	reschedule bool
	err        *error
}

// Defines an interface for a component responsible for some part of the
// reconciliation
type ControlFunction interface {
	// Sense - get information from the system/environment and update CR status
	// error -> log and move to the next CF
	Sense(spec *ar.ApicurioRegistry, request reconcile.Request) error

	// Compare the measured data in CR status with the intended state in CR spec,
	// Do not communicate with the environment here, just CR.
	// If true there was a discrepancy and the next stage will be executed, else skipped
	// or error -> log and move to next CF, reschedule
	Compare(spec *ar.ApicurioRegistry) (bool, error)

	// Do an action to get the system into alignment with the desired state
	// bool -> something was updated. reschedule until stable
	// error -> log and move to next CF, reschedule
	Respond(spec *ar.ApicurioRegistry) (bool, error)

	// Return the description of the CF
	Describe() string
}
