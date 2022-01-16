package config

import (
	"github.com/caarlos0/env/v6"
)

type confObj interface{}

func GetConfig(c confObj) error {
	err := env.Parse(c)
	if err != nil {
		return err
	}
	return nil
}
