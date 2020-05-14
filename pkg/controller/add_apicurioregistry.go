package controller

import (
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, apicurioregistry.Add)
}
