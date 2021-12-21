package storage

type Repositories interface {
	Connect() Repositories
	InsertGouge(name string, val float64)
	InsertCounter(name string, val int64)
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

func NewInMemoryDB() *InMemoryDB {
	return &InMemoryDB{Gouge: map[string]float64{}, Counter: map[string]int64{}}
}
