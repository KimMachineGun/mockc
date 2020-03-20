//+build mockc

package ex

import (
	"github.com/KimMachineGun/mockc"
)

func MockcCache() {
	mockc.Implement(Cache(nil))
}
