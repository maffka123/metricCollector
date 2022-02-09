package collector

import (
	"encoding/json"
	"fmt"
	"github.com/maffka123/metricCollector/internal/models"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"math/rand"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var runtimeMetricNameList = [...]string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction",
	"GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects",
	"HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
	"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC",
	"OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc"}

var psutilMetricNameList = [...]string{"TotalMemory", "FreeMemory"}

type Metric struct {
	Name     string
	prevVal  number
	currVal  number
	Change   number
	Type     string
	Key      *string
	memStats *runtime.MemStats
}

type PSMetric struct {
	Name    string
	prevVal number
	currVal number
	Change  number
	Type    string
	Key     *string
}

type MetricInterface interface {
	init()
	Print()
	Update(*sync.WaitGroup)
	MarshalJSON() ([]byte, error)
}

/*
GetAllMetrics prepares and intialize all metrics that are collected in this service.
see example here: https://github.com/tevjef/go-runtime-metrics/blob/master/collector/collector.go
*/
func GetAllMetrics(k *string) []*Metric {
	memStats := startStats()
	metricList := []*Metric{}
	for _, value := range runtimeMetricNameList {

		m := Metric{Name: value, Type: "gauge", Key: k, memStats: memStats}
		m.init()
		metricList = append(metricList, &m)
	}

	metricList = append(metricList, &Metric{Name: "PollCount", Type: "counter", currVal: number{integer: 1, float: 0.0}, Key: k})
	metricList = append(metricList, &Metric{Name: "RandomValue", Type: "gauge", currVal: number{integer: rand.Intn(100), float: 0.0}, Key: k})

	return metricList
}

// startStats prepares reference to runtime MemStats object
func startStats() *runtime.MemStats {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	return memStats
}

// init initializes starting value of the metric
func (m *Metric) init() {
	m.currVal.newNumber()
	m.prevVal.newNumber()
	m.Change.newNumber()

	m.MetricByName()
}

// MetricByName updates curr value of the metric using its name in memStats
func (m *Metric) MetricByName() {
	r := reflect.ValueOf(m.memStats)
	f := reflect.Indirect(r).FieldByName(m.Name)

	if f.Kind() == reflect.Float64 {
		m.currVal.float = f.Float()
	} else if f.Kind() == reflect.Uint64 {
		m.currVal.integer = int(f.Uint())
	}
}

//Print is more for debugging, print what is inside metric
func (m *Metric) Print() { fmt.Printf("%s: %d\n", m.Name, m.Change.Value()) }

// Update updates current value of the metric
func (m *Metric) Update(wg *sync.WaitGroup) {
	defer wg.Done()
	m.memStats = &runtime.MemStats{}
	runtime.ReadMemStats(m.memStats)
	m.prevVal = m.currVal
	if m.Name == "PollCount" {
		m.currVal.integer += 1
	} else if m.Name == "RandomValue" {
		m.currVal.integer = rand.Intn(100)
	} else {
		m.MetricByName()
	}

	m.Change = m.currVal.diff(&m.prevVal)
}

// MarshalJSON marshalls metrics to json
func (m *Metric) MarshalJSON() ([]byte, error) {
	newM := models.Metrics{}
	newM.ID = m.Name
	newM.MType = m.Type
	if newM.MType == "counter" {
		newM.Delta = m.Change.IntValue()
	} else if newM.MType == "gauge" {
		newM.Value = m.Change.FloatValue()
	}

	if m.Key != nil && *m.Key != "" {
		newM.CalcHash(*m.Key)
	}

	return json.Marshal(newM)
}

// GetAllPSUtilMetrics collects all psutil metrics at the start
func GetAllPSUtilMetrics(k *string) []*PSMetric {
	metricList := []*PSMetric{}
	for _, value := range psutilMetricNameList {

		m := PSMetric{Name: value, Type: "gauge", Key: k}
		m.init()
		metricList = append(metricList, &m)
	}

	psMetricCPU(&metricList, k)
	return metricList
}

// init initializes starting value of the metric
func (m *PSMetric) init() {
	m.currVal.newNumber()
	m.prevVal.newNumber()
	m.Change.newNumber()
	m.MetricByName()
}

// initWithVal initializes starting value of the metric using passed in value
func (m *PSMetric) initWithVal(val any) {
	m.currVal.newNumber()
	m.prevVal.newNumber()
	m.Change.newNumber()
	switch v := val.(type) {
	case float64:
		m.currVal.float = v
	case int64:
		m.currVal.integer = int(v)
	}
}

//Print is more for debugging, print what is inside metric
func (m *PSMetric) Print() { fmt.Printf("%s: %d\n", m.Name, m.Change.Value()) }

// MarshalJSON converts metrics to json
func (m *PSMetric) MarshalJSON() ([]byte, error) {
	newM := models.Metrics{}
	newM.ID = m.Name
	newM.MType = m.Type
	if newM.MType == "counter" {
		newM.Delta = m.Change.IntValue()
	} else if newM.MType == "gauge" {
		newM.Value = m.Change.FloatValue()
	}

	if m.Key != nil && *m.Key != "" {
		newM.CalcHash(*m.Key)
	}

	return json.Marshal(newM)
}

// Update updates metrics values
func (m *PSMetric) Update(wg *sync.WaitGroup) {
	defer wg.Done()

	m.prevVal = m.currVal
	if strings.Contains(m.Name, "CPUutilization") {
		j, _ := strconv.Atoi(m.Name[len(m.Name)-1:])
		c, _ := cpu.Times(true)

		m.currVal.float = c[j].Total()
	} else {
		m.MetricByName()

	}
	m.Change = m.currVal.diff(&m.prevVal)
}

// MetricByName finds metrics in psutil using their name
func (m *PSMetric) MetricByName() {
	v, _ := mem.VirtualMemory()

	if m.Name == "TotalMemory" {
		m.currVal.integer = int(v.Total)
	} else if m.Name == "FreeMemory" {
		m.currVal.integer = int(v.Free)
	}
}

// psMetricCPU collects cpu metrics for initialization of the metrics
func psMetricCPU(metricList *[]*PSMetric, k *string) {
	c, _ := cpu.Times(true)
	for i, usage := range c {
		m := PSMetric{Name: fmt.Sprintf("CPUutilization%d", i), Type: "gauge", Key: k}
		m.initWithVal(usage.Total())
		*metricList = append(*metricList, &m)
	}
}
