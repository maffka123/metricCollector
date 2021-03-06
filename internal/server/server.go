// Package server implements server
package server

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/maffka123/metricCollector/internal/server/config"
	"github.com/maffka123/metricCollector/internal/storage"
	"net/http"
)

func NewServer(endpoint string, handlers chi.Router) *http.Server {
	return &http.Server{Addr: endpoint, Handler: handlers}
}

/* DealWithDumps configures db dumping options: if store interval is >0 then it will be written asynchonousely
if store interavl is 0, dump will be triggered right after db change*/
func DealWithDumps(cfg *config.Config, db storage.Repositories, dbUpdated chan time.Time) {

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
func runDump(c <-chan time.Time, db storage.Repositories) {
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
