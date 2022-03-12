package storage

import (
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/maffka123/metricCollector/internal/models"
	"github.com/maffka123/metricCollector/internal/server/config"
)

// InMemoryDB type for holding all parameters of in-memory storage.
type InMemoryDB struct {
	Gouge         map[string]float64 `json:"gauge"`
	Counter       map[string]int64   `json:"counter"`
	StoreInterval time.Duration      `json:"-"`
	StoreFile     string             `json:"-"`
	Restore       bool               `json:"-"`
	log           *zap.Logger        `json:"-"`
}

// Connect initilizes in-memory storage.
func Connect(cfg *config.Config, logger *zap.Logger) *InMemoryDB {
	db := InMemoryDB{
		Gouge:         map[string]float64{},
		Counter:       map[string]int64{},
		StoreInterval: cfg.StoreInterval,
		StoreFile:     cfg.StoreFile,
		Restore:       cfg.Restore,
		log:           logger,
	}

	if cfg.Restore {
		err := db.RestoreDB()
		if err != nil {
			logger.Error("restore failed, starting with empty db: %s", zap.Error(errors.Unwrap(err)))
		}
	}

	return &db
}

// InsertGouge appends/updates gouge in metrics map.
func (db *InMemoryDB) InsertGouge(name string, val float64) {
	db.Gouge[name] = val
}

// InsertCounter appends/updates counter in metrics mao.
func (db *InMemoryDB) InsertCounter(name string, val int64) {
	db.Counter[name] += val
}

// NameInGouge checks if given gouge already exists in the map.
func (db *InMemoryDB) NameInGouge(s string) bool {
	if _, ok := db.Gouge[s]; ok {
		return true
	}
	return false
}

// NameInCounter checks if given counter already exists in the map.
func (db *InMemoryDB) NameInCounter(s string) bool {
	if _, ok := db.Counter[s]; ok {
		return true
	}
	return false
}

// ValueFromCounter gets counter by its name.
func (db *InMemoryDB) ValueFromCounter(s string) int64 {
	return db.Counter[s]
}

// ValueFromGouge gets gouge by its name.
func (db *InMemoryDB) ValueFromGouge(s string) float64 {
	return db.Gouge[s]
}

// SelectAll select all available metrics.
func (db *InMemoryDB) SelectAll() ([]string, []string) {
	var listCounter []string
	var listGouge []string

	for k, v := range db.Counter {
		listCounter = append(listCounter, fmt.Sprintf("[%s]: [%d]\n", k, v))
	}

	for k, v := range db.Gouge {
		listGouge = append(listGouge, fmt.Sprintf("[%s]: [%.3f]\n", k, v))
	}

	return listCounter, listGouge
}

// DumpDB stores metrics in a file as json.
func (db *InMemoryDB) DumpDB() error {
	p, err := NewProducer(db.StoreFile)

	if err != nil {
		db.log.Error("Producer initialisation failed")
		return err
	}
	defer p.Close()
	db.log.Info("Saved db")
	return p.encoder.Encode(&db)
}

// RestoreDB reads metrics from json file.
func (db *InMemoryDB) RestoreDB() error {
	c, err := NewConsumer(db.StoreFile)

	if err != nil {
		db.log.Error("Consumer initialisation failed")
		return err
	}
	defer c.Close()

	if err := c.decoder.Decode(&db); err != nil {
		return err
	}

	return nil
}

// CloseConnection empties metrics map.
func (db *InMemoryDB) CloseConnection() {
	var zeroDB = &InMemoryDB{}
	*db = *zeroDB
}

// BatchInsert insert several metrics at one time into map.
func (db *InMemoryDB) BatchInsert(ms []models.Metrics) {
	for _, m := range ms {
		if m.MType == "counter" {
			db.InsertCounter(m.ID, *m.Delta)
		} else {
			db.InsertGouge(m.ID, *m.Value)
		}
	}

}
