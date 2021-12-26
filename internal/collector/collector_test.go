package collector

import (
	"testing"

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
