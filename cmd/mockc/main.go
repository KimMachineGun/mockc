package main

import (
	"context"
	"log"
	"os"

	"github.com/KimMachineGun/mockc/internal/mockc"
)

func main() {
	log.SetFlags(0)

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("cannot get working directory:", err)
	}

	c := LoadConfig()
	if c.IsGeneratorMode() {
		err = mockc.Generate(context.Background(), wd, c.args)
	} else {
		err = c.ValidateFlags()
		if err == nil {
			err = mockc.GenerateWithFlags(context.Background(), wd, c.name, c.destination, c.args)
		}
	}
	if err != nil {
		log.Fatalln(err)
	}
}
