package storage

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/caarlos0/env/v6"
	"github.com/maffka123/metricCollector/internal/server/config"
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

func TestInMemoryDB_NameInGouge(t *testing.T) {
	type fields struct {
		Gouge   map[string]float64
		Counter map[string]int64
	}
	type args struct {
		s string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{name: "true", fields: fields{Gouge: map[string]float64{"g1": 0.5}, Counter: map[string]int64{"c1": 32}},
			args: args{s: "g1"}, want: true},
		{name: "false", fields: fields{Gouge: map[string]float64{"g1": 0.5}, Counter: map[string]int64{"c1": 32}},
			args: args{s: "g2"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &InMemoryDB{
				Gouge:   tt.fields.Gouge,
				Counter: tt.fields.Counter,
			}
			if got := db.NameInGouge(tt.args.s); got != tt.want {
				t.Errorf("InMemoryDB.NameInGouge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInMemoryDB_NameInCounter(t *testing.T) {
	type fields struct {
		Gouge   map[string]float64
		Counter map[string]int64
	}
	type args struct {
		s string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{name: "true", fields: fields{Gouge: map[string]float64{"g1": 0.5}, Counter: map[string]int64{"c1": 32}},
			args: args{s: "c1"}, want: true},
		{name: "false", fields: fields{Gouge: map[string]float64{"g1": 0.5}, Counter: map[string]int64{"c1": 32}},
			args: args{s: "c2"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &InMemoryDB{
				Gouge:   tt.fields.Gouge,
				Counter: tt.fields.Counter,
			}
			if got := db.NameInCounter(tt.args.s); got != tt.want {
				t.Errorf("InMemoryDB.NameInCounter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInMemoryDB_ValueFromCounter(t *testing.T) {
	type fields struct {
		Gouge   map[string]float64
		Counter map[string]int64
	}
	type args struct {
		s string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
	}{
		{name: "test1", fields: fields{Gouge: map[string]float64{"g1": 0.5}, Counter: map[string]int64{"c1": 32}},
			args: args{s: "c1"}, want: 32},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &InMemoryDB{
				Gouge:   tt.fields.Gouge,
				Counter: tt.fields.Counter,
			}
			if got := db.ValueFromCounter(tt.args.s); got != tt.want {
				t.Errorf("InMemoryDB.ValueFromCounter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInMemoryDB_ValueFromGouge(t *testing.T) {
	type fields struct {
		Gouge   map[string]float64
		Counter map[string]int64
	}
	type args struct {
		s string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   float64
	}{
		{name: "test1", fields: fields{Gouge: map[string]float64{"g1": 0.5}, Counter: map[string]int64{"c1": 32}},
			args: args{s: "g1"}, want: 0.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &InMemoryDB{
				Gouge:   tt.fields.Gouge,
				Counter: tt.fields.Counter,
			}
			if got := db.ValueFromGouge(tt.args.s); got != tt.want {
				t.Errorf("InMemoryDB.ValueFromGouge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInMemoryDB_SelectAll(t *testing.T) {
	type fields struct {
		Gouge   map[string]float64
		Counter map[string]int64
	}
	type want struct {
		GougeList   []string
		CounterList []string
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{name: "test1", fields: fields{Gouge: map[string]float64{"g1": 0.5}, Counter: map[string]int64{"c1": 32}},
			want: want{GougeList: []string{"[g1]: [0.500]\n"}, CounterList: []string{"[c1]: [32]\n"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &InMemoryDB{
				Gouge:   tt.fields.Gouge,
				Counter: tt.fields.Counter,
			}
			got, got1 := db.SelectAll()
			if !reflect.DeepEqual(got, tt.want.CounterList) {
				t.Errorf("InMemoryDB.SelectAll() got = %v, want %v", got, tt.want.CounterList)
			}
			if !reflect.DeepEqual(got1, tt.want.GougeList) {
				t.Errorf("InMemoryDB.SelectAll() got1 = %v, want %v", got1, tt.want.GougeList)
			}
		})
	}
}

func prepConf() *config.Config {
	var cfg config.Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	*cfg.Restore = false
	return &cfg
}

func TestInMemoryDB_DumpDB(t *testing.T) {
	cfg := prepConf()
	db := Connect(cfg)
	db.Counter["c1"] = 1
	db.Counter["c2"] = 2
	db.Gouge["g1"] = 1.5
	db.StoreFile = "testdata/dump.json"
	tests := []struct {
		name string
	}{
		{name: "test1"},
	}
	for _, tt := range tests {
		e := os.Remove(db.StoreFile)
		if e != nil {
			fmt.Println("file not found")
		}
		t.Run(tt.name, func(t *testing.T) {

			db.DumpDB()

			_, err := os.Stat(db.StoreFile)

			assert.NoError(t, err, os.ErrNotExist)

		})
	}
}

// This test is connected to the test above, it cheks the file created above
func TestInMemoryDB_RestoreDB(t *testing.T) {
	cfg := prepConf()
	db := Connect(cfg)
	db.StoreFile = "testdata/dump.json"
	type want struct {
		counter map[string]int64
		gauge   map[string]float64
	}
	tests := []struct {
		name string
		want want
	}{
		{name: "test1", want: want{counter: map[string]int64{"c1": 1, "c2": 2}, gauge: map[string]float64{"g1": 1.5}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := db.RestoreDB()

			assert.NoError(t, err)
			assert.Equal(t, tt.want.counter, db.Counter)
			assert.Equal(t, tt.want.gauge, db.Gouge)

		})
	}
}
