package agent

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"log"

	"github.com/caarlos0/env/v6"
	"github.com/maffka123/metricCollector/internal/agent/config"
	"github.com/maffka123/metricCollector/internal/collector"
	"github.com/stretchr/testify/assert"
)

func prepConf() *config.Config {
	var cfg config.Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	return &cfg
}

func Test_simpleBackoff(t *testing.T) {
	delay = 10 * time.Millisecond //s
	client := &http.Client{}
	timer := time.NewTimer(delay)
	cfg := prepConf()

	m := &collector.Metric{Name: "PollCount", Type: "counter"}
	fErr := sendDataFunc(func(cfg *config.Config, c *http.Client, m *collector.Metric) error {
		return errors.New("some error")
	})
	fNoerr := sendDataFunc(func(cfg *config.Config, c *http.Client, m *collector.Metric) error {
		select {
		case <-timer.C:
			return nil
		default:
			return errors.New("some error")
		}
	})

	type args struct {
		f sendDataFunc
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{name: "test1", args: args{f: fErr}, wantErr: errors.New("some error")},
		{name: "test2", args: args{f: fNoerr}, wantErr: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timer.Reset(delay)
			err := simpleBackoff(tt.args.f, cfg, client, m)
			assert.Equal(t, tt.wantErr, err)

		})
	}
}

func Test_sendData(t *testing.T) {
	client := &http.Client{}
	cfg := prepConf()

	type args struct {
		m *collector.Metric
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "test1", args: args{m: &collector.Metric{Name: "Alloc", Type: "gauge"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sendJSONData(cfg, client, tt.args.m)
			assert.Error(t, err)
		})
	}
}
