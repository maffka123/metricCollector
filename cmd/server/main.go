// server to collect metrics coming to its endpoints.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	globalConf "github.com/maffka123/metricCollector/internal/config"
	"github.com/maffka123/metricCollector/internal/handlers"
	"github.com/maffka123/metricCollector/internal/server"
	"github.com/maffka123/metricCollector/internal/server/config"
	"github.com/maffka123/metricCollector/internal/storage"
)

var (
	Version     string = "N/A"
	BuildDate   string = "N/A"
	BuildCommit string = "N/A"
)

// main implements all server logic.
// Shortly:
// - initialize config
// - initialize logger
// - initilize DB (can be in-memory of postgres)
// - initilize router
// - start goroutine to catch quit signal
// - start goroutine to make periodical db dumps
// - start serving
func main() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", Version, BuildDate, BuildCommit)
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}
	logger := globalConf.InitLogger(cfg.Debug)
	defer logger.Sync()

	logger.Info("Init config: done")

	var db storage.Repositories
	if cfg.DBpath != "" {
		db = storage.ConnectPG(context.Background(), &cfg, logger)
	} else {
		db = storage.Connect(&cfg, logger)
	}
	defer db.CloseConnection()

	logger.Info("Init config: done")

	r, dbUpdated := handlers.MetricRouter(db, &cfg, logger)

	srv := server.NewServer(cfg.Endpoint, r)

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// See example here: https://pkg.go.dev/net/http#example-Server.Shutdown
	// Gracefully Shutdown
	go func() {
		sig := <-quit
		logger.Info(fmt.Sprintf("caught sig: %+v", sig))
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			logger.Error("HTTP server Shutdown:", zap.Error(err))
		}
	}()

	go server.DealWithDumps(&cfg, db, dbUpdated)
	logger.Info("Start serving on", zap.String("endpoint name", cfg.Endpoint))
	log.Fatal(srv.ListenAndServe())

}
