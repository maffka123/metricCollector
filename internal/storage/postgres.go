package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/maffka123/metricCollector/internal/server/config"
	"os"
	"time"
)

type PGDB struct {
	StoreInterval time.Duration `json:"-"`
	StoreFile     string        `json:"-"`
	Restore       bool          `json:"-"`
	path          string        `json:"-"`
	Conn          *pgx.Conn     `json:"-"`
}

func ConnectPG(ctx context.Context, cfg *config.Config) *PGDB {
	db := PGDB{
		StoreInterval: cfg.StoreInterval,
		StoreFile:     cfg.StoreFile,
		Restore:       cfg.Restore,
		path:          cfg.DBpath,
	}
	conn, err := pgx.Connect(context.Background(), cfg.DBpath)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	db.Conn = conn

	db.Conn.Exec(ctx, "CREATE TABLE IF NOT EXISTS metrics (id serial PRIMARY KEY, name string UNIQUE NOT NULL, value float, type VARCHAR (10) NOT NULL);")

	/*if cfg.Restore {
		err := db.RestoreDB()
		if err != nil {
			fmt.Println(fmt.Errorf("restore failed: %s", err))
			fmt.Println("Restore failed, starting with empty db")
		}
	}*/

	return &db
}

func (db *PGDB) RestoreDB() error {
	return errors.New("not implemented")
}

func (db *PGDB) DumpDB() error {
	return errors.New("not implemented")
}

func (db *PGDB) SelectAll() ([]string, []string) {
	var listCounter []string
	var listGouge []string
	type counterRow struct {
		name  string
		value int64
	}
	type gaugeRow struct {
		name  string
		value float64
	}

	row, err := db.Conn.Query(context.Background(), "SELECT name, value FROM metrics WHERE type='counter'")
	if err != nil {
		fmt.Printf("Select counter failed: %s", err)
	}
	defer row.Close()

	for row.Next() {
		var r counterRow
		err := row.Scan(&r.name, &r.value)
		if err != nil {
			fmt.Printf("Select counter failed: %s", err)
		}
		listCounter = append(listCounter, fmt.Sprintf("[%s]: [%d]\n", r.name, r.value))
	}

	row, err = db.Conn.Query(context.Background(), "SELECT name, value FROM metrics WHERE type='gauge'")
	if err != nil {
		fmt.Printf("Select gauge failed: %s", err)
	}
	defer row.Close()

	for row.Next() {
		var r gaugeRow
		err := row.Scan(&r.name, &r.value)
		if err != nil {
			fmt.Printf("Select gauge failed: %s", err)
		}
		listGouge = append(listGouge, fmt.Sprintf("[%s]: [%.3f]\n", r.name, r.value))
	}
	return listCounter, listGouge
}

func (db *PGDB) InsertGouge(name string, val float64) {
	fmt.Println("not implemented")
}

func (db *PGDB) InsertCounter(name string, val int64) {
	fmt.Println("not implemented")
}

func (db *PGDB) NameInGouge(s string) bool {
	fmt.Println("not implemented")
	return false
}

func (db *PGDB) NameInCounter(s string) bool {
	fmt.Println("not implemented")
	return false
}

func (db *PGDB) ValueFromCounter(s string) int64 {
	var val int64
	row, err := db.Conn.Query(context.Background(), fmt.Sprintf("SELECT value FROM metrics WHERE name=%s", s))
	if err != nil {
		fmt.Printf("Select counter failed: %s", err)
	}
	defer row.Close()
	row.Scan(&val)
	return val
}

func (db *PGDB) ValueFromGouge(s string) float64 {
	var val float64
	row, err := db.Conn.Query(context.Background(), fmt.Sprintf("SELECT value FROM metrics WHERE name=%s", s))
	if err != nil {
		fmt.Printf("Select counter failed: %s", err)
	}
	defer row.Close()
	row.Scan(&val)
	return val
}
