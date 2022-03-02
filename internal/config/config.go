package config

import (
	"log"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
)

type confObj interface{}

func GetConfig(c confObj) error {
	err := env.Parse(c)
	if err != nil {
		return err
	}
	return nil
}

func InitLogger(debug bool) *zap.Logger {
	zapConfig := zap.NewProductionConfig()
	zapConfig.EncoderConfig.LevelKey = "severity"
	zapConfig.EncoderConfig.MessageKey = "message"

	if debug {
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	logger, err := zapConfig.Build(zap.Fields(
		zap.String("projectID", "MetricCollector"),
	))

	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}

	return logger
}
