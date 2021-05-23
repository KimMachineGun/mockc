//+build mockc

package basic

import (
	"fmt"

	"github.com/KimMachineGun/mockc"
)

func InvalidMockGenerator() {
	fmt.Println("Hello World")
	mockc.Implement(Cache(nil))
}
