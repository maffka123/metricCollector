package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/maffka123/metricCollector/internal/agent"
	"github.com/maffka123/metricCollector/internal/agent/config"
	internal "github.com/maffka123/metricCollector/internal/config"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var cfg config.Config

func main() {
	// for sending metrics to the server
	fmt.Println("starting the agent")

	flag.Parse()
	internal.GetConfig(&cfg)

	client := &http.Client{}

	// To be able to cancel tasks, make ctx and signal handler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	//starting list of metrices
	metricList := agent.InitMetrics(&cfg, client)

	//tickers for metric update and metric sending to the server
	pollTicker := time.NewTicker(cfg.PollInterval)
	reportTicker := time.NewTicker(cfg.ReportInterval)

	//start both tasks
	go agent.UpdateMetrics(ctx, pollTicker.C, metricList)
	go agent.SendAllData(ctx, &cfg, reportTicker.C, client, metricList)

	//if signal to qiut received cancel ctx
	<-quit
}

func init() {

	flag.StringVar(&cfg.Endpoint, "a", "127.0.0.1:8080", "server address as host:port")
	flag.DurationVar(&cfg.PollInterval, "p", 2*time.Second, "how often to update metrics")
	flag.DurationVar(&cfg.ReportInterval, "r", 10*time.Second, "how often to send metrics to the server")

}
