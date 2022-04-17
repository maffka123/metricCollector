// Package config holds configuration spezific for server.
package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/caarlos0/env/v6"
)

type rsaPrivKey rsa.PrivateKey

// Config type hold all configs for the server.
type Config struct {
	Endpoint      string        `env:"ADDRESS" json:"address"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" json:"store_interval"`
	StoreFile     string        `env:"STORE_FILE" json:"store_file"`
	Restore       bool          `env:"RESTORE" json:"restore"`
	Key           string        `env:"KEY"`
	DBpath        string        `env:"DATABASE_DSN" json:"database_dsn"`
	Debug         bool          `env:"METRIC_SERVER_DEBUG"`
	CryptoKey     rsaPrivKey    `env:"CRYPTO_KEY" json:"crypto_key"`
	configFile    string        `env:"CONFIG"`
	TrustedSubnet string        `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
}

func (v rsaPrivKey) String() string {
	return ""
}

func (v *rsaPrivKey) Set(s string) error {
	if s == "" {
		return nil
	}
	pub, err := ioutil.ReadFile(s)
	if err != nil {
		return err
	}
	block, _ := pem.Decode([]byte(pub))
	parseResult, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}

	*v = rsaPrivKey(*parseResult)
	return nil

}

// InitConfig allows to initialize config by first getting it from flags and then parsing environmental variables.
func InitConfig() (Config, error) {
	var cfg Config

	flag.StringVar(&cfg.Endpoint, "a", "localhost:8080", "server address as host:port")
	flag.BoolVar(&cfg.Restore, "r", true, "if to restore db from a dump")
	flag.DurationVar(&cfg.StoreInterval, "i", 300*time.Second, "how often to dump db into the file")
	flag.StringVar(&cfg.StoreFile, "f", "/tmp/devops-metrics-db.json", "name and location of the file path/to/file.json")
	flag.StringVar(&cfg.Key, "k", "", "key for hash function")
	flag.Var(&cfg.CryptoKey, "ck", "crypto key for asymmetric encoding")
	flag.BoolVar(&cfg.Debug, "debug", true, "key for hash function")
	flag.StringVar(&cfg.configFile, "c", "", "location of config.json file")
	flag.StringVar(&cfg.configFile, "t", "", "ip of a trusted agent")

	// find full options link here: https://github.com/jackc/pgx/blob/master/pgxpool/pool.go
	flag.StringVar(&cfg.DBpath, "d", "", "path for connection with pg: postgres://postgres:pass@localhost:5432/test?pool_max_conns=10")

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
		reflect.TypeOf(rsaPrivKey{}): rsaPrivKeyParser,
	}); err != nil {
		return err
	}
	return nil
}

func rsaPrivKeyParser(s string) (interface{}, error) {
	pub, err := ioutil.ReadFile(s)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pub)
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsaPrivKey(*privKey), nil
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
		StoreInterval string `json:"store_interval"`
		CryptoKey     string `json:"crypto_key"`
	}{
		// задаём указатель на целевой объект
		cAlias: (*cAlias)(c),
	}
	if err := json.Unmarshal(data, aliasValue); err != nil {
		return err
	}
	intUnits := aliasValue.StoreInterval[len(aliasValue.StoreInterval)-1:]
	if intUnits == "s" {
		intVar, err := strconv.Atoi(aliasValue.StoreInterval[:len(aliasValue.StoreInterval)-1])
		if err != nil {
			return err
		}
		c.StoreInterval = time.Duration(intVar) * time.Second
	} else {
		return errors.New("unknown time units")
	}

	var cryptoKey rsaPrivKey
	cryptoKey.Set(aliasValue.CryptoKey)
	c.CryptoKey = cryptoKey

	return nil
}
