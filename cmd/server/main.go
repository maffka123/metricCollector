package main

import (
	"context"
	"fmt"
	"github.com/maffka123/metricCollector/internal/handlers"
	"github.com/maffka123/metricCollector/internal/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var db = storage.NewInMemoryDB()

func main() {

	r := handlers.MetricRouter(db)
	srv := &http.Server{Addr: ":8080", Handler: r}

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

	fmt.Println("Start serving on localhost:8080")
	log.Fatal(srv.ListenAndServe())

}
