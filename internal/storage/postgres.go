package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/maffka123/metricCollector/internal/server/config"
	"os"
	"time"
)

type PGDB struct {
	StoreInterval time.Duration `json:"-"`
	StoreFile     string        `json:"-"`
	Restore       bool          `json:"-"`
	path          string        `json:"-"`
	Conn          *pgxpool.Pool `json:"-"`
}

func ConnectPG(ctx context.Context, cfg *config.Config) *PGDB {
	db := PGDB{
		StoreInterval: cfg.StoreInterval,
		StoreFile:     cfg.StoreFile,
		Restore:       cfg.Restore,
		path:          cfg.DBpath,
	}
	conn, err := pgxpool.Connect(context.Background(), cfg.DBpath)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		return &db
	}
	db.Conn = conn

	_, err = db.Conn.Exec(ctx, "CREATE TABLE IF NOT EXISTS metrics (id serial PRIMARY KEY, name VARCHAR (30) UNIQUE NOT NULL, value float, type VARCHAR (10) NOT NULL);")
	if err != nil {
		fmt.Printf("Table creation failed: %v\n", err)
	}

	if cfg.Restore {
		err := db.RestoreDB()
		if err != nil {
			fmt.Println(fmt.Errorf("restore failed: %s", err))
			fmt.Println("Restore failed, starting with empty db")
		}
	}

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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := db.Conn.Exec(ctx, `INSERT INTO metrics (name, value, type)
					VALUES($1,$2,'gauge') 
					ON CONFLICT (name) DO 
	    		UPDATE SET value = $2;`, name, val)
	if err != nil {
		fmt.Println(err)
	}
}

func (db *PGDB) InsertCounter(name string, val int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := db.Conn.Exec(ctx, `INSERT INTO metrics (name, value, type)
					VALUES($1,$2, 'counter') 
					ON CONFLICT (name) DO 
	    		UPDATE SET value = metrics.value+$2;`, name, val)
	if err != nil {
		fmt.Println(err)
	}
}

func (db *PGDB) NameInGouge(s string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var val float64
	row := db.Conn.QueryRow(ctx, "SELECT value FROM metrics WHERE name=$N", s)
	err := row.Scan(&val)
	return err == nil
}

func (db *PGDB) NameInCounter(s string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var val int64
	row := db.Conn.QueryRow(ctx, "SELECT value FROM metrics WHERE name=$1", s)
	err := row.Scan(&val)
	return err == nil
}

func (db *PGDB) ValueFromCounter(s string) int64 {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var val int64
	row := db.Conn.QueryRow(ctx, "SELECT value FROM metrics WHERE name=$1", s)
	err := row.Scan(&val)
	if err != nil {
		fmt.Printf("Select counter failed: %s", err)
	}

	return val
}

func (db *PGDB) ValueFromGouge(s string) float64 {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var val float64
	row := db.Conn.QueryRow(ctx, "SELECT value FROM metrics WHERE name=$1", s)
	err := row.Scan(&val)
	if err != nil {
		fmt.Printf("Select counter failed: %s", err)
	}

	return val
}

func (db *PGDB) CloseConnection() {
	db.Conn.Close()
}
