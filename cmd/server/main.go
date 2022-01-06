package main

import (
	"context"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/maffka123/metricCollector/internal/handlers"
	"github.com/maffka123/metricCollector/internal/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var db = storage.NewInMemoryDB()

type Config struct {
	Endpoint string `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
}

func main() {

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	r := handlers.MetricRouter(db)
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

	fmt.Printf("Start serving on %s\n", cfg.Endpoint)
	log.Fatal(srv.ListenAndServe())

}
