package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryDb_InsertCounter(t *testing.T) {
	type fields struct {
		Gouge   map[string]float64
		Counter map[string]int64
	}
	type args struct {
		name string
		val  int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
	}{
		{name: "new_val", fields: fields{Gouge: map[string]float64{"g1": 0.5}, Counter: map[string]int64{"c1": 32}},
			args: args{name: "Counter1", val: 1}, want: 1},
		{name: "increment", fields: fields{Gouge: map[string]float64{"g1": 0.5}, Counter: map[string]int64{"c1": 1}},
			args: args{name: "c1", val: 1}, want: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &InMemoryDB{
				Gouge:   tt.fields.Gouge,
				Counter: tt.fields.Counter,
			}
			db.InsertCounter(tt.args.name, tt.args.val)
			assert.Equal(t, tt.want, db.Counter[tt.args.name])
		})
	}
}

func TestInMemoryDb_InsertGouge(t *testing.T) {
	type fields struct {
		Gouge   map[string]float64
		Counter map[string]int64
	}
	type args struct {
		name string
		val  float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   float64
	}{
		{name: "replace", fields: fields{Gouge: map[string]float64{"g1": 0.5}, Counter: map[string]int64{"c1": 32}},
			args: args{name: "g1", val: 1.2}, want: 1.2},
		{name: "new_val", fields: fields{Gouge: map[string]float64{"g1": 0.5}, Counter: map[string]int64{"c1": 32}},
			args: args{name: "g11", val: 1.2}, want: 1.2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &InMemoryDB{
				Gouge:   tt.fields.Gouge,
				Counter: tt.fields.Counter,
			}
			db.InsertGouge(tt.args.name, tt.args.val)
			assert.Equal(t, tt.want, db.Gouge[tt.args.name])
		})
	}
}
