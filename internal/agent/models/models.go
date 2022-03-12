// agent models
package models

import (
	"github.com/maffka123/metricCollector/internal/collector"
)

// MetricList is a type used in agent to exchange metrics between its different parts.
type MetricList struct {
	MetricList []collector.MetricInterface
	Err        error
}
