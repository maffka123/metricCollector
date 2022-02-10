package main

import (
	"context"
	"errors"
	"github.com/maffka123/metricCollector/internal/agent"
	"github.com/maffka123/metricCollector/internal/agent/config"
	"github.com/maffka123/metricCollector/internal/agent/models"
	globalConf "github.com/maffka123/metricCollector/internal/config"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		panic(errors.Unwrap(err))
	}
}

func run() error {
	// for sending metrics to the server

	cfg, err := config.InitConfig()

	if err != nil {
		return err
	}

	logger := globalConf.InitLogger(cfg.Debug)
	defer logger.Sync()

	logger.Info("Agent config initialized")

	client := &http.Client{}

	// To be able to cancel tasks, make ctx and signal handler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	//starting list of metrices, with ability to cancel it
	var m models.MetricList
	ch := []chan models.MetricList{make(chan models.MetricList, 1), make(chan models.MetricList, 1)}
	outCh := make(chan models.MetricList, 1)
	go agent.InitMetrics(ctx, cfg, client, ch[0], logger)
	go agent.InitPSMetrics(ctx, cfg, client, ch[1], logger)
	go agent.FanIn(outCh, ch...)

metricList:
	for {
		select {
		case m = <-outCh:
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
	var cond sync.Mutex

	go agent.UpdateMetrics(ctx, cfg, &cond, pollTicker.C, m.MetricList, logger)
	go agent.SendAllData(ctx, cfg, &cond, reportTicker.C, client, m.MetricList, er, logger)
	logger.Info("Agent started")

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
