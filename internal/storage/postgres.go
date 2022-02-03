package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/maffka123/metricCollector/internal/models"
	"github.com/maffka123/metricCollector/internal/server/config"
	"go.uber.org/zap"
	"time"
)

type PGinterface interface {
	Begin(context.Context) (pgx.Tx, error)
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	Close()
}

type PGDB struct {
	StoreInterval time.Duration `json:"-"`
	StoreFile     string        `json:"-"`
	Restore       bool          `json:"-"`
	path          string        `json:"-"`
	Conn          PGinterface   `json:"-"`
	log           *zap.Logger   `json:"-"`
}

func ConnectPG(ctx context.Context, cfg *config.Config, logger *zap.Logger) *PGDB {
	db := PGDB{
		StoreInterval: cfg.StoreInterval,
		StoreFile:     cfg.StoreFile,
		Restore:       cfg.Restore,
		path:          cfg.DBpath,
		log:           logger,
	}
	conn, err := pgxpool.Connect(context.Background(), cfg.DBpath)

	if err != nil {
		db.log.Error("unable to connect to database: ", zap.Error(err))
		return &db
	}
	db.Conn = conn

	_, err = db.Conn.Exec(ctx, "CREATE TABLE IF NOT EXISTS metrics (id serial PRIMARY KEY, name VARCHAR (30) UNIQUE NOT NULL, value float, type VARCHAR (10) NOT NULL);")
	if err != nil {
		db.log.Error("table creation failed: ", zap.Error(err))
	}

	if cfg.Restore {
		err := db.RestoreDB()
		if err != nil {
			db.log.Error("restore failed, starting with empty db: ", zap.Error(err))
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
		db.log.Error("Select counter failed:", zap.Error(err))
	}
	defer row.Close()

	for row.Next() {
		var r counterRow
		err := row.Scan(&r.name, &r.value)
		if err != nil {
			db.log.Error("Select counter failed:", zap.Error(err))
		}
		listCounter = append(listCounter, fmt.Sprintf("[%s]: [%d]\n", r.name, r.value))
	}

	row, err = db.Conn.Query(context.Background(), "SELECT name, value FROM metrics WHERE type='gauge'")
	if err != nil {
		db.log.Error("Select gauge failed: %s", zap.Error(err))
	}
	defer row.Close()

	for row.Next() {
		var r gaugeRow
		err := row.Scan(&r.name, &r.value)
		if err != nil {
			db.log.Error("Select gauge failed: %s", zap.Error(err))
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
		db.log.Error("Insert gauge failed: ", zap.Error(err))
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
		db.log.Error("Insert counter failed: ", zap.Error(err))
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
		db.log.Error("select gauge failed: ", zap.Error(err))
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
		db.log.Error("select counter failed: ", zap.Error(err))
	}

	return val
}

func (db *PGDB) CloseConnection() {
	db.Conn.Close()
}

func (db *PGDB) BatchInsert(m []models.Metrics) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		db.log.Error("starting connection failed: ", zap.Error(err))
	}
	defer tx.Rollback(ctx)

	_, err = tx.Prepare(ctx, "batch insert counter", `INSERT INTO metrics (name, value, type)
													VALUES($1,$2,'counter') 
													ON CONFLICT (name) DO 
												UPDATE SET value = metrics.value+$2;`)

	if err != nil {
		db.log.Error("prep counter failed: ", zap.Error(err))
	}

	_, err = tx.Prepare(ctx, "batch insert gauge", `INSERT INTO metrics (name, value, type)
												VALUES($1,$2,'gauge') 
												ON CONFLICT (name) DO 
											UPDATE SET value = $2;`)
	if err != nil {
		db.log.Error("prep gauge failed: ", zap.Error(err))
	}

	for _, v := range m {
		if v.MType == "counter" {
			if _, err = tx.Exec(ctx, "batch insert counter", v.ID, v.Delta); err != nil {
				db.log.Error("Insert counter failed: ", zap.Error(err))
			}
		} else {
			if _, err = tx.Exec(ctx, "batch insert gauge", v.ID, v.Value); err != nil {
				db.log.Error("Insert gauge failed: ", zap.Error(err))
			}
		}

	}

	if err = tx.Commit(ctx); err != nil {
		db.log.Error("Commit failed: ", zap.Error(err))
	}

}
