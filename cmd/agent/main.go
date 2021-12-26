package main

import (
	"context"
	"fmt"
	"github.com/maffka123/metricCollector/internal/collector"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"
)

var endpoint string = "http://127.0.0.1:8080"
var pollInterval time.Duration = 2    //s
var reportInterval time.Duration = 10 //s

func main() {
	// for sending metrics to the server
	fmt.Println("starting the agent")

	_, err := http.Post("http://127.0.0.1:8080", "text/plain", nil)
	if err != nil {
		fmt.Println(err)
	}
	_, err1 := http.Post("http://localhost:8080", "text/plain", nil)
	if err1 != nil {
		fmt.Println(err1)
	}

	_, err2 := http.Post("http://127.0.0.1:8080/update/", "text/plain", nil)
	if err2 != nil {
		fmt.Println(err2)
	}

	_, err3 := http.Post("http://localhost:8080/update/counter/testCounter/100", "text/plain", nil)
	if err3 != nil {
		fmt.Println(err2)
	}

	client := &http.Client{}

	// To be able to cancel tasks, make ctx and signal handler
	ctx, cancel := context.WithCancel(context.Background())
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	//starting list of metrices
	metricList := initMetrics(client)

	//tickers for metric update and metric sending to the server
	pollTicker := time.NewTicker(pollInterval * time.Second)
	reportTicker := time.NewTicker(reportInterval * time.Second)

	//start both tasks
	go updateMetrics(ctx, pollTicker.C, metricList)
	go sendAllData(ctx, reportTicker.C, client, metricList)

	//if signal to qiut received cancel ctx
	<-quit
	cancel()
}

//initMetrics initializes list with all metrics of interest, send first values to the server
func initMetrics(client *http.Client) []*collector.Metric {
	metricList := collector.GetAllMetrics()
	for _, value := range metricList {
		value.Print()
		sendData(client, value)
	}
	return metricList
}

//updateMetrics updates metrics from the list
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
			fmt.Println("context canceled")
		}
	}
}

//sendData sends metric to the server.
func sendData(client *http.Client, m *collector.Metric) {
	var url string
	if reflect.TypeOf(m.Change.Value()).Kind() == reflect.Int64 || reflect.TypeOf(m.Change.Value()).Kind() == reflect.Int {
		url = fmt.Sprintf("%s/update/%s/%s/%d", endpoint, m.Type, m.Name, m.Change.Value())
	} else {
		url = fmt.Sprintf("%s/update/%s/%s/%f", endpoint, m.Type, m.Name, m.Change.Value())
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
	defer response.Body.Close()

	fmt.Println("Sent data with status code", response.Status)

}

// sendAllData iterates over metrics list and sent them to the server
func sendAllData(ctx context.Context, t <-chan time.Time, client *http.Client, metricList []*collector.Metric) {
	for {

		select {
		case <-t:
			fmt.Println("Sending all metrics")
			for _, value := range metricList {
				sendData(client, value)
			}
		case <-ctx.Done():
			fmt.Println("context canceled")
		}
	}
}
