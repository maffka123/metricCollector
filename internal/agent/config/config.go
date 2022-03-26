// Package config holds specific config for agent
package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"io/ioutil"
	"reflect"
	"time"

	"github.com/caarlos0/env/v6"
)

// Config is a majoj config structure.
type Config struct {
	Endpoint       string        `env:"ADDRESS"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	Retries        int           `env:"BACKOFF_RETRIES"`
	Delay          time.Duration `env:"BACKOFF_DELAY"`
	Key            string        `env:"KEY"`
	Debug          bool          `env:"METRIC_SERVER_DEBUG"`
	Profile        bool          `env:"METRIC_SERVER_PROFILE"`
	CryptoKey      rsaPubKey     `env:"CRYPTO_KEY"`
}

type rsaPubKey rsa.PublicKey

func (v *rsaPubKey) String() string {
	return ""
}

func (v *rsaPubKey) Set(s string) error {
	if s == "" {
		return nil
	}
	pub, err := ioutil.ReadFile(s)
	if err != nil {
		return err
	}
	pubPem, _ := pem.Decode(pub)
	pubkey, err := x509.ParsePKCS1PublicKey(pubPem.Bytes)
	if err != nil {
		return err
	}

	*v = rsaPubKey(*pubkey)
	return nil
}

// InitConfig initilizes config so that it first checks flags and the env variables.
func InitConfig() (Config, error) {
	var cfg Config

	flag.StringVar(&cfg.Endpoint, "a", "127.0.0.1:8080", "server address as host:port")
	flag.DurationVar(&cfg.PollInterval, "p", 2*time.Second, "how often to update metrics")
	flag.DurationVar(&cfg.ReportInterval, "r", 10*time.Second, "how often to send metrics to the server")
	flag.IntVar(&cfg.Retries, "n", 3, "how many times should try to send metrics in case of error")
	flag.DurationVar(&cfg.Delay, "t", 10*time.Second, "delay in case of error and retry")
	flag.StringVar(&cfg.Key, "k", "", "key for hash function")
	flag.Var(&cfg.CryptoKey, "ck", "crypto key for asymmetric encoding")
	flag.BoolVar(&cfg.Debug, "debug", true, "if debugging is needed")
	flag.BoolVar(&cfg.Profile, "profile", false, "if profiling is needed")

	flag.Parse()
	err := GetConfig(&cfg)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func GetConfig(cfg *Config) error {
	if err := env.ParseWithFuncs(cfg, map[reflect.Type]env.ParserFunc{
		reflect.TypeOf(rsaPubKey{}): rsaPubKeyParser,
	}); err != nil {
		return err
	}
	return nil
}

func rsaPubKeyParser(s string) (interface{}, error) {
	pub, err := ioutil.ReadFile(s)
	if err != nil {
		return nil, err
	}
	pubPem, _ := pem.Decode(pub)
	pubkey, err := x509.ParsePKCS1PublicKey(pubPem.Bytes)
	if err != nil {
		return nil, err
	}
	return rsaPubKey(*pubkey), nil
}
