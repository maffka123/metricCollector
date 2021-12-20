package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maffka123/metricCollector/internal/collector"
)

var endpoint string = "http://localhost:8080"
var pollInterval time.Duration = 2    //s
var reportInterval time.Duration = 10 //s

func main() {
	client := &http.Client{}
	//metricListChan := make(chan []collector.Metric)

	metricList := initMetrics(client)

	ctx, cancel := context.WithCancel(context.Background())

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	pollTicker := time.NewTicker(pollInterval * time.Second)
	reportTicker := time.NewTicker(reportInterval * time.Second)

	go updateMetrics(ctx, pollTicker.C, metricList)
	go sendAllData(ctx, reportTicker.C, client, metricList)

	<-quit
	cancel()
}

func initMetrics(client *http.Client) []*collector.Metric {
	metricList := collector.GetAllMetrics()
	for _, value := range metricList {
		value.Print()
		sendData(client, value)
	}
	return metricList
}

func updateMetrics(ctx context.Context, t <-chan time.Time, metricList []*collector.Metric) {
	for {
		select {
		case <-t:
			fmt.Println("Updating all metrics")
			for _, value := range metricList {
				value.Update()
				value.Print()
			}
		case <-ctx.Done():
			fmt.Errorf("context canceled")
		}
	}
}

func sendData(client *http.Client, m *collector.Metric) {
	url := fmt.Sprintf("%s/update/%s/%s/%d", endpoint, m.Type, m.Name, m.Change.Value())
	if m.Change.Value() != 0 {
		fmt.Println(url)
	}

	request, err := http.NewRequest(http.MethodPost, url, nil)
	request.Header.Add("application-type", "text/plain")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	response, err2 := client.Do(request)
	if err2 != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Sent data with status code", response.Status)
	defer response.Body.Close()
}

func sendAllData(ctx context.Context, t <-chan time.Time, client *http.Client, metricList []*collector.Metric) {
	for {

		select {
		case <-t:
			fmt.Println("Sending all metrics")
			for _, value := range metricList {
				sendData(client, value)
			}
		case <-ctx.Done():
			fmt.Errorf("context canceled")
		}
	}
}
