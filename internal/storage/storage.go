package storage

import "fmt"

type Repositories interface {
	Connect() Repositories
	InsertGouge(name string, val float64)
	InsertCounter(name string, val int64)
	NameInCounter(s string) bool
	NameInGouge(s string) bool
	ValueFromCounter(s string) int64
	ValueFromGouge(s string) float64
	SelectAll() ([]string, []string)
}

type InMemoryDB struct {
	Gouge   map[string]float64
	Counter map[string]int64
}

func (db *InMemoryDB) Connect() Repositories {
	return db
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

func NewInMemoryDB() *InMemoryDB {
	return &InMemoryDB{Gouge: map[string]float64{}, Counter: map[string]int64{}}
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
