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

var db = storage.NewInMemoryDb()

func main() {

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	http.HandleFunc("/update/gouge/", handlers.PostHandlerGouge(db))
	http.HandleFunc("/update/count/", handlers.PostHandlerCounter(db))
	fmt.Println("Start serving on localhost:8080")
	go func() {
		sig := <-quit
		fmt.Printf("caught sig: %+v", sig)
		os.Exit(0)
	}()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
