package main

import (
	"context"
	"fmt"
	"github.com/maffka123/metricCollector/internal/handlers"
	"github.com/maffka123/metricCollector/internal/server"
	"github.com/maffka123/metricCollector/internal/server/config"
	"github.com/maffka123/metricCollector/internal/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.InitConfig()

	db := storage.Connect(&cfg)
	pg := storage.ConnectPG(context.Background(), &cfg)

	r, dbUpdated := handlers.MetricRouter(db, &cfg.Key)
	r.Get("/ping", handlers.GetHandlerPing(pg))
	r.Get("/", handlers.GetAllNames(db))

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
