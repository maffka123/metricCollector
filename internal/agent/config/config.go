// Package config holds specific config for agent
package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"time"

	"encoding/json"
	"github.com/caarlos0/env/v6"
)

// Config is a majoj config structure.
type Config struct {
	Endpoint       string        `env:"ADDRESS" json:"address"`
	EndpointGRPC   string        `env:"ADDRESS_GRPC" json:"address_grpc"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" json:"report_interval"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" json:"poll_interval"`
	Retries        int           `env:"BACKOFF_RETRIES"`
	Delay          time.Duration `env:"BACKOFF_DELAY"`
	Key            string        `env:"KEY"`
	Debug          bool          `env:"METRIC_SERVER_DEBUG"`
	Profile        bool          `env:"METRIC_SERVER_PROFILE"`
	CryptoKey      rsaPubKey     `env:"CRYPTO_KEY" json:"crypto_key"`
	configFile     string        `env:"CONFIG"`
	Protocol       string        `env:"PROTOCOL"`
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
	flag.StringVar(&cfg.EndpointGRPC, "ag", "localhost:8082", "server address as host:port")
	flag.DurationVar(&cfg.PollInterval, "p", 2*time.Second, "how often to update metrics")
	flag.DurationVar(&cfg.ReportInterval, "r", 10*time.Second, "how often to send metrics to the server")
	flag.IntVar(&cfg.Retries, "n", 3, "how many times should try to send metrics in case of error")
	flag.DurationVar(&cfg.Delay, "t", 10*time.Second, "delay in case of error and retry")
	flag.StringVar(&cfg.Key, "k", "", "key for hash function")
	flag.Var(&cfg.CryptoKey, "ck", "crypto key for asymmetric encoding")
	flag.BoolVar(&cfg.Debug, "debug", true, "if debugging is needed")
	flag.BoolVar(&cfg.Profile, "profile", false, "if profiling is needed")
	flag.StringVar(&cfg.configFile, "c", "", "location of config.json file")
	flag.StringVar(&cfg.Protocol, "pr", "http", "send data over http or grpc")

	// config from env variables
	flag.Parse()

	// config from flags
	if err := GetConfig(&cfg); err != nil {
		return cfg, err
	}

	// config from json file
	if cfg.configFile != "" {
		if err := parseConfigFile(&cfg); err != nil {
			return cfg, err
		}
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

func parseConfigFile(cfg *Config) error {
	jsonFile, err := os.Open(cfg.configFile)
	if err != nil {
		return err
	}

	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	if err := json.Unmarshal(byteValue, cfg); err != nil {
		return err
	}
	return nil
}

func (c *Config) UnmarshalJSON(data []byte) error {
	type cAlias Config

	aliasValue := &struct {
		*cAlias
		// переопределяем поле внутри анонимной структуры
		ReportInterval string `json:"report_interval"`
		PollInterval   string `json:"poll_interval"`
		CryptoKey      string `json:"crypto_key"`
	}{
		// задаём указатель на целевой объект
		cAlias: (*cAlias)(c),
	}
	if err := json.Unmarshal(data, aliasValue); err != nil {
		return err
	}
	intUnits := aliasValue.ReportInterval[len(aliasValue.ReportInterval)-1:]
	if intUnits == "s" {
		intVar, err := strconv.Atoi(aliasValue.ReportInterval[:len(aliasValue.ReportInterval)-1])
		if err != nil {
			return err
		}
		c.ReportInterval = time.Duration(intVar) * time.Second
	} else {
		return errors.New("unknown time units")
	}

	intUnits = aliasValue.PollInterval[len(aliasValue.PollInterval)-1:]
	if intUnits == "s" {
		intVar, err := strconv.Atoi(aliasValue.PollInterval[:len(aliasValue.PollInterval)-1])
		if err != nil {
			return err
		}
		c.PollInterval = time.Duration(intVar) * time.Second
	} else {
		return errors.New("unknown time units")
	}

	var cryptoKey rsaPubKey
	cryptoKey.Set(aliasValue.CryptoKey)
	c.CryptoKey = cryptoKey

	return nil
}
