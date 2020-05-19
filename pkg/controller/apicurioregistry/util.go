// Some code in this file was adopted from https://github.com/atlasmap/atlasmap-operator
package apicurioregistry

import (
	"github.com/go-logr/logr"
)

func fatal(log logr.Logger, err error, msg string) {
	log.Error(err, msg)
	panic("Fatal error, the operator can't recover.")
}
