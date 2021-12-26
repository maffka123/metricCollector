package main

import (
	"fmt"
	"github.com/maffka123/metricCollector/cmd/server/handlers"
	"github.com/maffka123/metricCollector/cmd/server/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var db = storage.NewInMemoryDB()

func main() {

	r := handlers.MetricRouter(db)

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		sig := <-quit
		fmt.Printf("caught sig: %+v", sig)
		os.Exit(0)
	}()

	fmt.Println("Start serving on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))

}
