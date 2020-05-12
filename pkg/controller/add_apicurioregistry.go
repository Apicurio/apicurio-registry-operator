package controller

import (
	"github.com/apicurio/apicurio-operators/apicurio-registry/pkg/controller/apicurioregistry"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, apicurioregistry.Add)
}
