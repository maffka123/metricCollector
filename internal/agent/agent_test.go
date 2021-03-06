package agent

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"log"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/maffka123/metricCollector/internal/agent/config"
	"github.com/maffka123/metricCollector/internal/collector"
)

var logger *zap.Logger = zap.NewExample()

func prepConf() config.Config {
	var cfg config.Config
	err := config.GetConfig(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}

func Test_simpleBackoff(t *testing.T) {
	delay := 10 * time.Millisecond //s
	client := &http.Client{}
	timer := time.NewTimer(delay)
	cfg := prepConf()
	cfg.Delay = delay
	cfg.Retries = 3
	ctx := context.Background()

	m := []*collector.Metric{{Name: "PollCount", Type: "counter"}}
	fErr := sendDataFunc(func(ctx context.Context, cfg config.Config, c *http.Client, m []collector.MetricInterface, logger *zap.Logger) error {
		return errors.New("some error")
	})
	fNoerr := sendDataFunc(func(ctx context.Context, cfg config.Config, c *http.Client, m []collector.MetricInterface, logger *zap.Logger) error {
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
			a := make([]collector.MetricInterface, len(m))
			for i := range m {
				a[i] = m[i]
			}
			timer.Reset(delay)
			err := simpleBackoff(ctx, tt.args.f, cfg, client, a, logger)
			assert.Equal(t, tt.wantErr, err)

		})
	}
}

func Test_sendData(t *testing.T) {

	client := &http.Client{}

	ctx := context.Background()

	type args struct {
		m []*collector.Metric
	}
	type settings struct {
		retries   string
		key       string
		pathtokey string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		settings settings
	}{

		{name: "test1", args: args{m: []*collector.Metric{{Name: "Count1", Type: "counter"}}},
			settings: settings{retries: "3", key: "test", pathtokey: "config/testdata/key.pem"},
		},
		{name: "test2", args: args{m: []*collector.Metric{{Name: "Count1", Type: "counter"}}},
			settings: settings{retries: "3", key: "test", pathtokey: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("CRYPTO_KEY", tt.settings.pathtokey)
			os.Setenv("BACKOFF_RETRIES", tt.settings.retries)
			os.Setenv("KEY", tt.settings.key)
			cfg := prepConf()

			m := make([]collector.MetricInterface, len(tt.args.m))
			for i := range tt.args.m {
				tt.args.m[i].Key = &cfg.Key
				m[i] = tt.args.m[i]
			}
			err := sendJSONData(ctx, cfg, client, m, logger)
			assert.Error(t, err)
		})
	}
}
