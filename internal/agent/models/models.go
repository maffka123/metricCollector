package models

import (
	"github.com/maffka123/metricCollector/internal/collector"
)

type MetricList struct {
	MetricList []*collector.Metric
	Err        error
}