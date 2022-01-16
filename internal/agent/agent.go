package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/maffka123/metricCollector/internal/agent/config"
	"github.com/maffka123/metricCollector/internal/agent/models"
	"github.com/maffka123/metricCollector/internal/collector"
	"net/http"
	"time"
)

type sendDataFunc func(context.Context, config.Config, *http.Client, *collector.Metric) error

//initMetrics initializes list with all metrics of interest, send first values to the server
func InitMetrics(ctx context.Context, cfg config.Config, client *http.Client, ch chan models.MetricList) {
	//ctx, cancel := context.WithCancel(ctx)
	metricList := collector.GetAllMetrics()
	for _, value := range metricList {
		select {
		case <-ctx.Done():
			fmt.Println("context canceled")
		default:
			err := simpleBackoff(ctx, sendJSONData, cfg, client, value)
			if err != nil {
				ch <- models.MetricList{MetricList: nil, Err: err}
			}
		}
	}
	ch <- models.MetricList{MetricList: metricList, Err: nil}
}

//updateMetrics updates metrics from the list
func UpdateMetrics(ctx context.Context, t <-chan time.Time, metricList []*collector.Metric) {
	for {
		select {
		case <-t:
			fmt.Println("Updating all metrics")
			for _, value := range metricList {
				value.Update()
				value.Print()
			}
		case <-ctx.Done():
			fmt.Println("context canceled")
		}
	}
}

//sendJSONData sends metric in json format to the server.
func sendJSONData(ctx context.Context, cfg config.Config, client *http.Client, m *collector.Metric) error {
	url := fmt.Sprintf("http://%s/update/", cfg.Endpoint)

	metricToSend, err := json.Marshal(m)
	if err != nil {
		fmt.Println(err)
		return err
	}

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(metricToSend))
	request.Header.Add("Content-Type", "application/json")

	if err != nil {
		fmt.Println(err)
		return err
	}

	response, requestErr := client.Do(request)
	if requestErr != nil {
		fmt.Println(requestErr)
		return requestErr
	}
	defer response.Body.Close()

	fmt.Println("Sent data with status code", response.Status)
	return nil
}

// sendAllData iterates over metrics list and sent them to the server
func SendAllData(ctx context.Context, cfg config.Config, t <-chan time.Time, client *http.Client, metricList []*collector.Metric) {
	ctx, cancel := context.WithCancel(ctx)
	for {

		select {
		case <-t:
			fmt.Println("Sending all metrics")
			for _, value := range metricList {
				err := simpleBackoff(ctx, sendJSONData, cfg, client, value)
				if err != nil {
					cancel()
				}
			}
		case <-ctx.Done():
			fmt.Println("context canceled")
		}
	}
}

// simpleBackoff repeats call to a function in case of an error
func simpleBackoff(ctx context.Context, f sendDataFunc, cfg config.Config, c *http.Client, m *collector.Metric) error {
	var err error
	for i := 0; i < cfg.Retries; i++ {
		select {
		case <-ctx.Done():
			fmt.Println("context canceled")
			return nil
		default:
			err = f(ctx, cfg, c, m)
			if err == nil {
				break
			}
			fmt.Printf("Backing off number %d\n", i+1)
			time.Sleep(cfg.Delay * time.Duration(i+1))
		}
	}
	return err
}
