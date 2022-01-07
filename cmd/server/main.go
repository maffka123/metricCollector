package main

import (
	"context"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/maffka123/metricCollector/internal/handlers"
	"github.com/maffka123/metricCollector/internal/server"
	"github.com/maffka123/metricCollector/internal/server/models"
	"github.com/maffka123/metricCollector/internal/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	var cfg models.Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

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
