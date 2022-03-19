// Package config holds configuration spezific for server.
package config

import (
	"flag"
	"time"

	internal "github.com/maffka123/metricCollector/internal/config"
)

// Config type hold all configs for the server.
type Config struct {
	Endpoint      string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
	Key           string        `env:"KEY"`
	DBpath        string        `env:"DATABASE_DSN"`
	Debug         bool          `env:"METRIC_SERVER_DEBUG"`
}

// InitConfig allows to initialize config by first getting it from flags and then parsing environmental variables.
func InitConfig() Config {
	var cfg Config

	flag.StringVar(&cfg.Endpoint, "a", "localhost:8080", "server address as host:port")
	flag.BoolVar(&cfg.Restore, "r", true, "if to restore db from a dump")
	flag.DurationVar(&cfg.StoreInterval, "i", 300*time.Second, "how often to dump db into the file")
	flag.StringVar(&cfg.StoreFile, "f", "/tmp/devops-metrics-db.json", "name and location of the file path/to/file.json")
	flag.StringVar(&cfg.Key, "k", "", "key for hash function")
	flag.BoolVar(&cfg.Debug, "debug", true, "key for hash function")

	// find full options link here: https://github.com/jackc/pgx/blob/master/pgxpool/pool.go
	flag.StringVar(&cfg.DBpath, "d", "", "path for connection with pg: postgres://postgres:pass@localhost:5432/test?pool_max_conns=10")

	flag.Parse()
	internal.GetConfig(&cfg)

	return cfg
}
