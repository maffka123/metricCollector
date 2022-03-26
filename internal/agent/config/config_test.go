// Package config holds configuration spezific for agent.
package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	type settings struct {
		Endpoint        string
		pathToCryptoKey string
	}
	type want struct {
		Endpoint  string
		CryptoKey int
	}
	tests := []struct {
		name     string
		settings settings
		want     want
	}{
		{name: "parse key", settings: settings{Endpoint: "localhost:8080", pathToCryptoKey: "testdata/key.pem"},
			want: want{Endpoint: "localhost:8080", CryptoKey: 65537}},
		{name: "parse empty key", settings: settings{Endpoint: "", pathToCryptoKey: ""},
			want: want{Endpoint: "", CryptoKey: 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg Config

			os.Setenv("ADDRESS", tt.settings.Endpoint)
			os.Setenv("CRYPTO_KEY", tt.settings.pathToCryptoKey)
			GetConfig(&cfg)
			assert.Equal(t, cfg.Endpoint, tt.want.Endpoint)
			assert.Equal(t, cfg.CryptoKey.E, tt.want.CryptoKey)
		})
	}
}

func TestFlagForKeys(t *testing.T) {
	type settings struct {
		Endpoint        string
		pathToCryptoKey string
	}
	type want struct {
		Endpoint  string
		CryptoKey int
	}
	tests := []struct {
		name     string
		settings settings
		want     want
	}{
		{name: "parse key", settings: settings{Endpoint: "localhost:8080", pathToCryptoKey: "testdata/key.pem"},
			want: want{Endpoint: "localhost:8080", CryptoKey: 65537}},
		{name: "parse empty key", settings: settings{Endpoint: "", pathToCryptoKey: ""},
			want: want{Endpoint: "", CryptoKey: 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg Config
			os.Args = []string{"cmd", "-a", tt.settings.Endpoint, "-ck", tt.settings.pathToCryptoKey}
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			flag.StringVar(&cfg.Endpoint, "a", "localhost:8080", "server address as host:port")
			flag.Var(&cfg.CryptoKey, "ck", "crypto key for asymmetric encoding")
			flag.Parse()
			assert.Equal(t, cfg.Endpoint, tt.want.Endpoint)
			assert.Equal(t, cfg.CryptoKey.E, tt.want.CryptoKey)
		})
	}
}
