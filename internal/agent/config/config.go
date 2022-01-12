package config

import (
	"flag"
	internal "github.com/maffka123/metricCollector/internal/config"
	"time"
)

type Config struct {
	Endpoint       string        `env:"ADDRESS"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
}

func InitConfig() Config {
	var cfg Config

	flag.StringVar(&cfg.Endpoint, "a", "127.0.0.1:8080", "server address as host:port")
	flag.DurationVar(&cfg.PollInterval, "p", 2*time.Second, "how often to update metrics")
	flag.DurationVar(&cfg.ReportInterval, "r", 10*time.Second, "how often to send metrics to the server")

	flag.Parse()
	internal.GetConfig(&cfg)
	return cfg
}
