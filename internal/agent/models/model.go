package models

import "time"

type Config struct {
	Endpoint       string        `env:"ADDRESS" envDefault:"http://127.0.0.1:8080"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
}
