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
var retries int = 3
var delay time.Duration = 10 //s
type sendDataFunc func(*http.Client, *collector.Metric) error

func main() {
	// for sending metrics to the server
	fmt.Println("starting the agent")

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
		//sendData(client, value)
		err := simpleBackoff(sendData, client, value)
		if err != nil {
			os.Exit(1)
		}
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
func sendData(client *http.Client, m *collector.Metric) error {
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
func sendAllData(ctx context.Context, t <-chan time.Time, client *http.Client, metricList []*collector.Metric) {
	for {

		select {
		case <-t:
			fmt.Println("Sending all metrics")
			for _, value := range metricList {
				simpleBackoff(sendData, client, value)
			}
		case <-ctx.Done():
			fmt.Println("context canceled")
		}
	}
}

//go does not see it when in another file
// simpleBackoff repeats call to a function in case of an error
func simpleBackoff(f sendDataFunc, c *http.Client, m *collector.Metric) error {
	var err error
	for i := 0; i < retries; i++ {
		err = f(c, m)
		if err == nil {
			break
		}
		fmt.Printf("Backing off number %d\n", i+1)
		time.Sleep(time.Second * delay * time.Duration(i+1))
	}
	return err
}
