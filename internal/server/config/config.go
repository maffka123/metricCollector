package config

import (
	"flag"
	internal "github.com/maffka123/metricCollector/internal/config"
	"time"
)

type Config struct {
	Endpoint      string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
	Key           string        `env:"KEY"`
	DBpath        string        `env:"DATABASE_DSN"`
}

func InitConfig() Config {
	var cfg Config

	flag.StringVar(&cfg.Endpoint, "a", "localhost:8080", "server address as host:port")
	flag.BoolVar(&cfg.Restore, "r", true, "if to restore db from a dump")
	flag.DurationVar(&cfg.StoreInterval, "i", 300*time.Second, "how often to dump db into the file")
	flag.StringVar(&cfg.StoreFile, "f", "/tmp/devops-metrics-db.json", "name and location of the file path/to/file.json")
	flag.StringVar(&cfg.Key, "k", "", "key for hash function")
	flag.StringVar(&cfg.DBpath, "d", "", "path for connection with pg: postgres://postgres:pass@localhost:5432/test")

	flag.Parse()
	internal.GetConfig(&cfg)

	return cfg
}
