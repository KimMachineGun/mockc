//+build mockc

package ex

import (
	"github.com/KimMachineGun/mockc"
)

func MockcCache() {
	mockc.Implements(Cache(nil))
}
