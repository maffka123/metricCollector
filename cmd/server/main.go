package main

import (
	"context"
	"fmt"

	"flag"
	internal "github.com/maffka123/metricCollector/internal/config"
	"github.com/maffka123/metricCollector/internal/handlers"
	"github.com/maffka123/metricCollector/internal/server"
	"github.com/maffka123/metricCollector/internal/server/config"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maffka123/metricCollector/internal/storage"
)

var cfg config.Config

func main() {

	flag.Parse()
	internal.GetConfig(&cfg)

	db := storage.Connect(&cfg)

	r, dbUpdated := handlers.MetricRouter(db)
	srv := &http.Server{Addr: cfg.Endpoint, Handler: r}

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	//See example here: https://pkg.go.dev/net/http#example-Server.Shutdown
	go func() {
		sig := <-quit
		fmt.Printf("caught sig: %+v", sig)
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
	}()

	go server.DealWithDumps(&cfg, db, dbUpdated)

	fmt.Printf("Start serving on %s\n", cfg.Endpoint)
	log.Fatal(srv.ListenAndServe())

}

func init() {

	flag.StringVar(&cfg.Endpoint, "a", "127.0.0.1:8080", "server address as host:port")
	flag.BoolVar(&cfg.Restore, "r", true, "if to restore db from a dump")
	flag.DurationVar(&cfg.StoreInterval, "i", 300*time.Second, "how often to dump db into the file")
	flag.StringVar(&cfg.StoreFile, "f", "/tmp/devops-metrics-db.json", "name and location of the file path/to/file.json")

}
