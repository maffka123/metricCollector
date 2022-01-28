package collector

import (
	"encoding/json"
	"testing"

	"github.com/maffka123/metricCollector/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestMetric_Update(t *testing.T) {
	type fields struct {
		Name    string
		prevVal number
		currVal number
		Change  number
		Type    string
	}
	type wait struct {
		prevVal number
		currVal number
		Change  number
	}
	tests := []struct {
		name   string
		fields fields
		wait   wait
	}{
		{name: "start", fields: fields{Name: "PollCount", currVal: number{integer: 1}, Type: "count"},
			wait: wait{prevVal: number{integer: 1}, currVal: number{integer: 2}, Change: number{integer: 1}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metric{
				Name:    tt.fields.Name,
				prevVal: tt.fields.prevVal,
				currVal: tt.fields.currVal,
				Change:  tt.fields.Change,
				Type:    tt.fields.Type,
			}
			m.Update()
			assert.Equal(t, tt.wait.Change.Value(), m.Change.Value())
			assert.Equal(t, tt.wait.prevVal.Value(), m.prevVal.Value())
		})
	}
}

func TestGetAllMetrics(t *testing.T) {
	key := "test"
	tests := []struct {
		name string
		want []*Metric
	}{
		{name: "test1", want: []*Metric{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAllMetrics(&key)
			assert.IsType(t, tt.want, got)
			assert.Equal(t, 30, len(got))
		})
	}
}

func TestMetric_init(t *testing.T) {
	memStats := startStats()
	type fields struct {
		Name string
		Type string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{name: "test1", fields: fields{Name: "Alloc", Type: "gauge"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metric{
				Name: tt.fields.Name,
				Type: tt.fields.Type,
			}
			m.init(memStats)
			assert.Equal(t, 0, m.prevVal.integer)
			assert.Equal(t, 0, m.Change.integer)
			assert.NotEqual(t, 0, m.currVal.integer)
		})
	}
}

func TestMetric_MarshalJSON(t *testing.T) {
	f := float64(1)
	type fields struct {
		Name   string
		Change number
		Type   string
	}
	tests := []struct {
		name   string
		fields fields
		want   models.Metrics
		key    string
	}{
		{name: "test1", fields: fields{Name: "Alloc", Change: number{integer: 1}, Type: "gauge"},
			want: models.Metrics{ID: "Alloc", MType: "gauge", Value: &f, Hash: "8bc975450597ab163a49d8ee05461ef1b8d7734ca97e1c9f0bf3f691f68a0a11"},
			key:  "test",
		},
		{name: "test1", fields: fields{Name: "Alloc", Change: number{integer: 1}, Type: "gauge"},
			want: models.Metrics{ID: "Alloc", MType: "gauge", Value: &f},
			key:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metric{
				Name:   tt.fields.Name,
				Change: tt.fields.Change,
				Type:   tt.fields.Type,
				Key:    &tt.key,
			}
			got, err := m.MarshalJSON()
			assert.NoError(t, err)
			want, _ := json.Marshal(tt.want)
			assert.Equal(t, got, want)
		})
	}
}
