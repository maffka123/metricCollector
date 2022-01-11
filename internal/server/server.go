package server

import (
	"github.com/maffka123/metricCollector/internal/server/config"
	"github.com/maffka123/metricCollector/internal/storage"
	"time"
)

/* DealWithDumps configures db dumping options: if store interval is >0 then it will be written asynchonousely
if store interavl is 0, dump will be triggered right after db change*/
func DealWithDumps(cfg *config.Config, db *storage.InMemoryDB, dbUpdated chan time.Time) {

	if cfg.StoreFile != "" && cfg.StoreInterval != 0 {
		storeTicker := time.NewTicker(cfg.StoreInterval)
		go runDump(storeTicker.C, db)
		go flushChannel(dbUpdated)
	} else if cfg.StoreFile != "" && cfg.StoreInterval == 0 {
		go runDump(dbUpdated, db)
	} else {
		go flushChannel(dbUpdated)
	}

}

// runDump starts infinite loop to dump db asynchronesely
func runDump(c <-chan time.Time, db *storage.InMemoryDB) {
	for {
		<-c
		db.DumpDB()
	}
}

//flushChannel allows to free channel
func flushChannel(c <-chan time.Time) {
	for {
		<-c
	}
}
