//+build mockc

package basic

import (
	"github.com/KimMachineGun/mockc"
)

func MockcCache() {
	mockc.Implement("Cache(nil)")
}
