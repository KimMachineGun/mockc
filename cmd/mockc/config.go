package main

import (
	"errors"
	"flag"
)

type Config struct {
	name        string
	destination string
	args        []string
}

func (c Config) IsGeneratorMode() bool {
	if c.name != "" || c.destination != "" {
		return false
	}

	return true
}

func (c Config) ValidateFlags() error {
	if c.name == "" {
		return errors.New("name flag is required in command line flags mode")
	}
	if c.destination == "" {
		return errors.New("destination flag is required in command line flags mode")
	}

	return nil
}

func LoadConfig() Config {
	var c Config

	flag.StringVar(&c.name, "name", "", "flag mode: name of generated mock")
	flag.StringVar(&c.destination, "destination", "", "flag mode: destination of generated file")

	flag.Parse()

	c.args = flag.Args()

	return c
}
