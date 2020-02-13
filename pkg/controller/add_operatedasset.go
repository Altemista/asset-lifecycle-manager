package controller

import (
	"github.com/Altemista/asset-lifecycle-manager/pkg/controller/operatedasset"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, operatedasset.Add)
}
