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

	"github.com/maffka123/metricCollector/internal/storage"
)

var cfg config.Config
var envCfg config.Config

func main() {

	flag.Parse()

	fmt.Println(cfg.Endpoint)

	db := storage.Connect(&cfg)

	r, dbUpdated := handlers.MetricRouter(db)
	srv := &http.Server{Addr: *cfg.Endpoint, Handler: r}

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

	fmt.Printf("Start serving on %s\n", *cfg.Endpoint)
	log.Fatal(srv.ListenAndServe())

}

func init() {

	internal.GetConfig(&envCfg)

	cfg.Endpoint = flag.String("a", *envCfg.Endpoint, "server address as host:port")
	cfg.Restore = flag.Bool("r", *envCfg.Restore, "if to restore db from a dump")
	cfg.StoreInterval = flag.Duration("i", *envCfg.StoreInterval, "how often to dump db into the file")
	cfg.StoreFile = flag.String("f", *envCfg.StoreFile, "name and location of the file path/to/file.json")
}
