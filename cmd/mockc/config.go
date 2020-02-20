package main

import (
	"errors"
	"flag"
)

type Config struct {
	name            string
	destination     string
	fieldNamePrefix string
	fieldNameSuffix string
	args            []string
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
	if c.fieldNamePrefix == "" && c.fieldNameSuffix == "" {
		return errors.New("at least one of the fieldNamePrefix and fieldNameSuffix must not be an empty string")
	}

	return nil
}

func LoadConfig() Config {
	var c Config

	flag.StringVar(&c.name, "name", "", "flag mode: name of the mock")
	flag.StringVar(&c.destination, "destination", "", "flag mode: mock file destination")
	flag.StringVar(&c.fieldNamePrefix, "fieldNamePrefix", "_", "flag mode: prefix of the mock's field names")
	flag.StringVar(&c.fieldNameSuffix, "fieldNameSuffix", "", "flag mode: suffix of the mock's field names")

	flag.Parse()

	c.args = flag.Args()

	return c
}
