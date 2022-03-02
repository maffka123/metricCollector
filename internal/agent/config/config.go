package config

import (
	"flag"
	"time"

	internal "github.com/maffka123/metricCollector/internal/config"
)

type Config struct {
	Endpoint       string        `env:"ADDRESS"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	Retries        int           `env:"BACKOFF_RETRIES"`
	Delay          time.Duration `env:"BACKOFF_DELAY"`
	Key            string        `env:"KEY"`
	Debug          bool          `env:"METRIC_SERVER_DEBUG"`
	Profile        bool          `env:"METRIC_SERVER_PROFILE"`
}

func InitConfig() (Config, error) {
	var cfg Config

	flag.StringVar(&cfg.Endpoint, "a", "127.0.0.1:8080", "server address as host:port")
	flag.DurationVar(&cfg.PollInterval, "p", 2*time.Second, "how often to update metrics")
	flag.DurationVar(&cfg.ReportInterval, "r", 10*time.Second, "how often to send metrics to the server")
	flag.IntVar(&cfg.Retries, "n", 3, "how many times should try to send metrics in case of error")
	flag.DurationVar(&cfg.Delay, "t", 10*time.Second, "delay in case of error and retry")
	flag.StringVar(&cfg.Key, "k", "", "key for hash function")
	flag.BoolVar(&cfg.Debug, "debug", true, "if debugging is needed")
	flag.BoolVar(&cfg.Profile, "profile", true, "if profiling is needed")

	flag.Parse()
	err := internal.GetConfig(&cfg)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}
