//+build mockc

package basic

import (
	"github.com/KimMachineGun/mockc"
)

func MockcTypeCode() {
	mockc.Implement(TypeCode(nil))
}
