package mockc

import (
	"go/types"
)

type Mode int

type Mock struct {
	Name       string
	Mode       Mode
	Interfaces []*types.Interface
}
