package agent

import (
	"net/http"
	"testing"

	"github.com/maffka123/metricCollector/internal/collector"
)

func Test_simpleBackoff(t *testing.T) {
	type args struct {
		f sendDataFunc
		c *http.Client
		m *collector.Metric
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := simpleBackoff(tt.args.f, tt.args.c, tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("simpleBackoff() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
