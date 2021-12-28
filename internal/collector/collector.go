package collector

import (
	"fmt"
	"math/rand"
	"runtime"
)

var runtimeMetricNameList = [...]string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction",
	"GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects",
	"HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
	"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC",
	"OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys"}

type Metric struct {
	Name    string
	prevVal number
	currVal number
	Change  number
	Type    string
}

/*
GetAllMetrics prepares and intialize all metrics that are collected in this service.
see example here: https://github.com/tevjef/go-runtime-metrics/blob/master/collector/collector.go
*/
func GetAllMetrics() []*Metric {
	memStats := startStats()
	metricList := []*Metric{}
	for _, value := range runtimeMetricNameList {

		m := Metric{Name: value, Type: "gauge"}
		m.init(memStats)
		metricList = append(metricList, &m)
	}

	metricList = append(metricList, &Metric{Name: "PollCount", Type: "counter", currVal: number{integer: 1, float: 0.0}})
	metricList = append(metricList, &Metric{Name: "RandomValue", Type: "gauge", currVal: number{integer: rand.Intn(100), float: 0.0}})

	return metricList
}

// startStats prepares reference to runtime MemStats object
func startStats() *runtime.MemStats {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	return memStats
}

// init initializes starting value of the metric
func (m *Metric) init(memStats *runtime.MemStats) {
	m.currVal.newNumber()
	m.prevVal.newNumber()
	m.Change.newNumber()

	runtimeMetricByName(m, memStats)
}

//Print is more for debugging, print what is inside metric
func (m *Metric) Print() { fmt.Printf("%s: %d\n", m.Name, m.Change.Value()) }

func (m *Metric) Update() {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	m.prevVal = m.currVal
	if m.Name == "PollCount" {
		m.currVal.integer += 1
	} else if m.Name == "RandomValue" {
		m.currVal.integer = rand.Intn(100)
	} else {
		runtimeMetricByName(m, memStats)
	}

	m.Change = m.currVal.diff(&m.prevVal)
}
