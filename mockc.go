package mockc

import (
	. "github.com/KimMachineGun/mockc/internal/mockc"
)

const (
	Basic Mode = 1 << iota
)

func SetMode(Mode)              {}
func Implements(...interface{}) {}
