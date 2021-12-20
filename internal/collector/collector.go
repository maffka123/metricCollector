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

func (n *number) newNumber() {
	n.float = 0
	n.integer = 0
}

type Metric struct {
	Name    string
	prevVal number
	currVal number
	Change  number
	Type    string
}

/*
Prepare and intialize all metrics that are collected in this service.
see example here: https://github.com/tevjef/go-runtime-metrics/blob/master/collector/collector.go
*/
func GetAllMetrics() []Metric {
	memStats := startStats()
	metricList := []Metric{}
	for _, value := range runtimeMetricNameList {

		m := Metric{Name: value, Type: "gauge"}
		m.init(memStats)
		metricList = append(metricList, m)
	}

	metricList = append(metricList, Metric{Name: "PollCount", Type: "count", currVal: number{integer: 1, float: 0}})
	metricList = append(metricList, Metric{Name: "RandomValue", Type: "gouge", currVal: number{integer: rand.Intn(100), float: 0}})

	return metricList
}

// Prepare reference to runtime MemStats object
func startStats() *runtime.MemStats {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	return memStats
}

// Initialize starting value of the metric
func (m *Metric) init(memStats *runtime.MemStats) {
	m.currVal.newNumber()
	m.prevVal.newNumber()
	m.Change.newNumber()

	runtimeMetricByName(m, memStats)
}

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
