package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/KimMachineGun/mockc/internal/mockc"
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("cannot get working directory:", err)
	}

	m := mockc.NewMockc(wd, flag.Args())

	err = m.Execute(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
}
