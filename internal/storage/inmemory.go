package storage

import (
	"fmt"
	"time"

	"github.com/maffka123/metricCollector/internal/server/config"
)

type InMemoryDB struct {
	Gouge         map[string]float64 `json:"gauge"`
	Counter       map[string]int64   `json:"counter"`
	StoreInterval time.Duration      `json:"-"`
	StoreFile     string             `json:"-"`
	Restore       bool               `json:"-"`
}

func Connect(cfg *config.Config) *InMemoryDB {
	db := InMemoryDB{
		Gouge:         map[string]float64{},
		Counter:       map[string]int64{},
		StoreInterval: *cfg.StoreInterval,
		StoreFile:     *cfg.StoreFile,
		Restore:       *cfg.Restore}

	if *cfg.Restore {
		err := db.RestoreDB()
		if err != nil {
			fmt.Println(fmt.Errorf("restore failed: %s", err))
			fmt.Println("Restore failed, starting with empty db")
		}
	}

	return &db
}

func (db *InMemoryDB) InsertGouge(name string, val float64) {
	db.Gouge[name] = val
}

func (db *InMemoryDB) InsertCounter(name string, val int64) {
	db.Counter[name] += val
}

func (db *InMemoryDB) NameInGouge(s string) bool {
	if _, ok := db.Gouge[s]; ok {
		return true
	}
	return false
}

func (db *InMemoryDB) NameInCounter(s string) bool {
	if _, ok := db.Counter[s]; ok {
		return true
	}
	return false
}

func (db *InMemoryDB) ValueFromCounter(s string) int64 {
	return db.Counter[s]
}

func (db *InMemoryDB) ValueFromGouge(s string) float64 {
	return db.Gouge[s]
}

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

func (db *InMemoryDB) DumpDB() error {
	p, err := NewProducer(db.StoreFile)

	if err != nil {
		fmt.Println("Producer initialisation failed")
		return err
	}
	defer p.Close()
	fmt.Println("Saved db")
	return p.encoder.Encode(&db)
}

func (db *InMemoryDB) RestoreDB() error {
	c, err := NewConsumer(db.StoreFile)

	if err != nil {
		fmt.Println("Consumer initialisation failed")
		return err
	}
	defer c.Close()

	if err := c.decoder.Decode(&db); err != nil {
		return err
	}

	return nil
}