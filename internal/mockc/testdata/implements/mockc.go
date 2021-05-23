//+build mockc

package basic

import (
	"github.com/KimMachineGun/mockc"
)

func MockcCache() {
	mockc.Implements(Cache(nil))
}
