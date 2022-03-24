// Package storage holds all what is connected to data storage.
package storage

import (
	"github.com/maffka123/metricCollector/internal/models"
)

// Repositories interface to allow using different databases.
type Repositories interface {
	InsertGouge(name string, val float64)
	InsertCounter(name string, val int64)
	NameInCounter(s string) bool
	NameInGouge(s string) bool
	ValueFromCounter(s string) int64
	ValueFromGouge(s string) float64
	SelectAll() ([]string, []string)
	DumpDB() error
	RestoreDB() error
	CloseConnection()
	BatchInsert([]models.Metrics)
}
