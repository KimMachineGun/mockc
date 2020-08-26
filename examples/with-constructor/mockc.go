//+build mockc

package constructor

import (
	"github.com/KimMachineGun/mockc"
)

func MockcCache() {
	mockc.Implement(Cache(nil))
	mockc.WithConstructor()
}
