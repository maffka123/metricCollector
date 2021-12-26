package collector

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_number_diff(t *testing.T) {
	type fields struct {
		integer int
		float   float64
	}
	type args struct {
		m *number
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   number
	}{
		{name: "int_positive", fields: fields{integer: 2, float: 0.0}, args: args{&number{integer: 1, float: 0.0}}, want: number{integer: 1, float: 0.0}},
		{name: "int_neg", fields: fields{integer: 1, float: 0.0}, args: args{&number{integer: 2, float: 0.0}}, want: number{integer: -1, float: 0.0}},
		{name: "float_positive", fields: fields{integer: 0, float: 2.0}, args: args{&number{integer: 0, float: 1.0}}, want: number{integer: 0, float: 1.0}},
		{name: "float_neg", fields: fields{integer: 0, float: 1.0}, args: args{&number{integer: 0, float: 2.0}}, want: number{integer: 0, float: -1.0}},
		{name: "zero", fields: fields{integer: 0, float: 0.0}, args: args{&number{integer: 0, float: 0.0}}, want: number{integer: 0, float: 0.0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &number{
				integer: tt.fields.integer,
				float:   tt.fields.float,
			}
			if got := n.diff(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("number.diff() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_number_newNumber(t *testing.T) {
	type want struct {
		integer int
		float   float64
	}
	tests := []struct {
		name string
		want want
	}{
		{name: "test1", want: want{integer: 0, float: 0.0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &number{}
			n.newNumber()
			assert.IsType(t, &number{}, n)
			assert.Equal(t, tt.want.integer, n.integer)
			assert.Equal(t, tt.want.float, n.float)

		})
	}
}
