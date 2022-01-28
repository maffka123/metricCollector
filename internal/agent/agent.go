package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/maffka123/metricCollector/internal/agent/config"
	"github.com/maffka123/metricCollector/internal/agent/models"
	"github.com/maffka123/metricCollector/internal/collector"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type sendDataFunc func(context.Context, config.Config, *http.Client, []*collector.Metric, *zap.Logger) error

//initMetrics initializes list with all metrics of interest, send first values to the server
func InitMetrics(ctx context.Context, cfg config.Config, client *http.Client, ch chan models.MetricList, logger *zap.Logger) {
	//ctx, cancel := context.WithCancel(ctx)
	metricList := collector.GetAllMetrics(&cfg.Key)

	err := simpleBackoff(ctx, sendJSONData, cfg, client, metricList, logger)
	if err != nil {
		ch <- models.MetricList{MetricList: nil, Err: err}
	}
	a := models.MetricList{MetricList: metricList, Err: nil}
	ch <- a
}

//updateMetrics updates metrics from the list
func UpdateMetrics(ctx context.Context, t <-chan time.Time, metricList []*collector.Metric, logger *zap.Logger) {
	for {
		select {
		case <-t:
			logger.Info("Updating all metrics")
			for _, value := range metricList {
				value.Update()
			}
		case <-ctx.Done():
			logger.Info("context canceled")
		}
	}
}

//sendJSONData sends metric in json format to the server.
func sendJSONData(ctx context.Context, cfg config.Config, client *http.Client, m []*collector.Metric, logger *zap.Logger) error {
	url := fmt.Sprintf("http://%s/updates/", cfg.Endpoint)

	metricToSend, err := json.Marshal(m)

	if err != nil {
		logger.Error("JSON marshal failed", zap.Error(err))
		return err
	}

	// gzip data
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write(metricToSend)
	gz.Close()

	// create a request
	request, err := http.NewRequest(http.MethodPost, url, &buf)
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Content-Encoding", "gzip")

	if err != nil {
		logger.Error("request creation failed", zap.Error(err))
		return err
	}

	// execute the request
	response, requestErr := client.Do(request)
	if requestErr != nil {
		fmt.Println(requestErr)
		return requestErr
	}
	defer response.Body.Close()

	logger.Info("Sent data with status code", zap.String("code", response.Status))
	return nil
}

// sendAllData iterates over metrics list and sent them to the server
func SendAllData(ctx context.Context, cfg config.Config, t <-chan time.Time, client *http.Client, metricList []*collector.Metric, er chan error, logger *zap.Logger) {

	for {

		select {
		case <-t:
			fmt.Println("Sending all metrics")
			err := simpleBackoff(ctx, sendJSONData, cfg, client, metricList, logger)
			if err != nil {
				er <- err
			}
		case <-ctx.Done():
			fmt.Println("context canceled")
		}
	}
}

// simpleBackoff repeats call to a function in case of an error
func simpleBackoff(ctx context.Context, f sendDataFunc, cfg config.Config, c *http.Client, m []*collector.Metric, logger *zap.Logger) error {
	var err error
backoff:
	for i := 0; i < cfg.Retries; i++ {
		select {
		case <-ctx.Done():
			logger.Info("context canceled")
			return nil
		default:
			err = f(ctx, cfg, c, m, logger)
			if err == nil {
				break backoff
			}
			logger.Info("Backing off number", zap.String("n", string(i+1)))
			time.Sleep(cfg.Delay * time.Duration(i+1))
		}
	}
	return err
}
