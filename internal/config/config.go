package config

import (
	"github.com/caarlos0/env/v6"
	"log"
)

type confObj interface{}

func GetConfig(c confObj) {
	err := env.Parse(c)
	if err != nil {
		log.Fatal(err)
	}

}
