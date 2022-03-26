// Package config package usefull for both serivces here. It allows initialize configs in the same way for both.
package config

import (
	"log"

	"go.uber.org/zap"
)

// InitLogger initilizes and configures zap logger
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
