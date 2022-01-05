package collector

import (
	"reflect"
	"runtime"
)

type any interface{}

type number struct {
	integer int
	float   float64
}

func (n *number) Value() any {

	if n.integer != 0 {
		return n.integer
	} else if n.float != 0.0 {
		return n.float
	} else {
		return 0
	}

}

func (n *number) FloatValue() *float64 {
	var f float64
	if n.integer != 0 {
		f = float64(n.integer)
	} else if n.float != 0.0 {
		f = n.float
	} else {
		f = 0.0
	}
	return &f

}

func (n *number) IntValue() *int64 {
	var i int64
	if n.integer != 0 {
		i = int64(n.integer)
	} else {
		i = 0
	}
	return &i
}

func (n *number) diff(m *number) number {
	if n.integer != 0 {
		return number{integer: n.integer - m.integer, float: 0}
	} else if n.float != 0.0 {
		return number{float: n.float - m.float, integer: 0}
	} else {
		return number{integer: 0, float: 0}
	}
}

func runtimeMetricByName(m *Metric, memStats *runtime.MemStats) {
	r := reflect.ValueOf(memStats)
	f := reflect.Indirect(r).FieldByName(m.Name)

	if f.Kind() == reflect.Float64 {
		m.currVal.float = f.Float()
	} else if f.Kind() == reflect.Uint64 {
		m.currVal.integer = int(f.Uint())
	}
}

func (n *number) newNumber() {
	n.float = 0.0
	n.integer = 0
}
