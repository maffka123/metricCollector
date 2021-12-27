package main

import (
	"context"
	"fmt"
	"github.com/maffka123/metricCollector/internal/agent"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var pollInterval time.Duration = 2 * time.Second    //s
var reportInterval time.Duration = 10 * time.Second //s

func main() {
	// for sending metrics to the server
	fmt.Println("starting the agent")

	client := &http.Client{}

	// To be able to cancel tasks, make ctx and signal handler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	//starting list of metrices
	metricList := agent.InitMetrics(client)

	//tickers for metric update and metric sending to the server
	pollTicker := time.NewTicker(pollInterval)
	reportTicker := time.NewTicker(reportInterval)

	//start both tasks
	go agent.UpdateMetrics(ctx, pollTicker.C, metricList)
	go agent.SendAllData(ctx, reportTicker.C, client, metricList)

	//if signal to qiut received cancel ctx
	<-quit
}
