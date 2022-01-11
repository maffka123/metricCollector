package config

import (
	"testing"

	"github.com/maffka123/metricCollector/internal/server/config"
	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	type args struct {
		c config.Config
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "test1", args: args{c: config.Config{Endpoint: "test.com:4565"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GetConfig(&tt.args.c)
			assert.Equal(t, tt.args.c.Endpoint, "test.com:4565")
			assert.Equal(t, tt.args.c.StoreFile, "")
		})
	}
}
