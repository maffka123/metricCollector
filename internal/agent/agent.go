package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/maffka123/metricCollector/internal/agent/config"
	"github.com/maffka123/metricCollector/internal/collector"
)

var retries int = 3
var delay time.Duration = 10 * time.Second //s
type sendDataFunc func(*config.Config, *http.Client, *collector.Metric) error

//initMetrics initializes list with all metrics of interest, send first values to the server
func InitMetrics(cfg *config.Config, client *http.Client) []*collector.Metric {
	metricList := collector.GetAllMetrics()
	for _, value := range metricList {
		value.Print()
		//sendData(client, value)
		err := simpleBackoff(sendJSONData, cfg, client, value)
		if err != nil {
			os.Exit(1)
		}
	}
	return metricList
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
func sendJSONData(cfg *config.Config, client *http.Client, m *collector.Metric) error {
	url := fmt.Sprintf("http://%s/update/", cfg.Endpoint)

	metricToSend, err := json.Marshal(m)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
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
func SendAllData(ctx context.Context, cfg *config.Config, t <-chan time.Time, client *http.Client, metricList []*collector.Metric) {
	for {

		select {
		case <-t:
			fmt.Println("Sending all metrics")
			for _, value := range metricList {
				simpleBackoff(sendJSONData, cfg, client, value)
			}
		case <-ctx.Done():
			fmt.Println("context canceled")
		}
	}
}

// simpleBackoff repeats call to a function in case of an error
func simpleBackoff(f sendDataFunc, cfg *config.Config, c *http.Client, m *collector.Metric) error {
	var err error
	for i := 0; i < retries; i++ {
		err = f(cfg, c, m)
		if err == nil {
			break
		}
		fmt.Printf("Backing off number %d\n", i+1)
		time.Sleep(delay * time.Duration(i+1))
	}
	return err
}
