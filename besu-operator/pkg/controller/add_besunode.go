package controller

import (
	"github.com/Sumaid/besu-kubernetes/besu-operator/pkg/controller/besunode"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, besunode.Add)
}
