package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	type config struct {
		Endpoint  string `env:"ADDRESS"`
		StoreFile string `env:"STORE_FILE"`
	}
	type args struct {
		c config
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "test1", args: args{c: config{Endpoint: "test.com:4565"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GetConfig(&tt.args.c)
			assert.Equal(t, tt.args.c.Endpoint, "test.com:4565")
			assert.Equal(t, tt.args.c.StoreFile, "")
		})
	}
}
