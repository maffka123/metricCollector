package main

import (
	"context"
	"fmt"
	"github.com/maffka123/metricCollector/internal/agent"
	"github.com/maffka123/metricCollector/internal/agent/config"
	"github.com/maffka123/metricCollector/internal/agent/models"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		panic("unexpected error: " + err.Error())
	}
}

func run() error {
	// for sending metrics to the server
	fmt.Println("starting the agent")
	cfg, err := config.InitConfig()

	if err != nil {
		return err
	}

	client := &http.Client{}

	// To be able to cancel tasks, make ctx and signal handler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	//starting list of metrices, with ability to cancel it
	var m models.MetricList
	ch := make(chan models.MetricList, 1)
	go agent.InitMetrics(ctx, cfg, client, ch)

metricList:
	for {
		select {
		case m = <-ch:
			if m.Err != nil {
				return m.Err
			}
			break metricList
		case <-quit:
			return nil
		}
	}

	//tickers for metric update and metric sending to the server
	pollTicker := time.NewTicker(cfg.PollInterval)
	reportTicker := time.NewTicker(cfg.ReportInterval)

	//start both tasks
	var er chan error
	go agent.UpdateMetrics(ctx, pollTicker.C, m.MetricList)
	go agent.SendAllData(ctx, cfg, reportTicker.C, client, m.MetricList, er)

	//if signal to qiut or error from other functions received, cancel ctx
	for {
		select {
		case <-quit:
			return nil
		case err = <-er:
			return err
		}
	}

}
