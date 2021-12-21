package storage

type Repositories interface {
	Connect() Repositories
	InsertGouge(name string, val float64)
	InsertCounter(name string, val int64)
}

type InMemoryDb struct {
	Gouge   map[string]float64
	Counter map[string]int64
}

func (db *InMemoryDb) Connect() Repositories {
	return db
}

func (db *InMemoryDb) InsertGouge(name string, val float64) {
	db.Gouge[name] = val
}

func (db *InMemoryDb) InsertCounter(name string, val int64) {
	db.Counter[name] += val
}

func NewInMemoryDb() *InMemoryDb {
	return &InMemoryDb{Gouge: map[string]float64{}, Counter: map[string]int64{}}
}
