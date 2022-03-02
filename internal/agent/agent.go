// agent implemets util methods spezific for agent.

package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/maffka123/metricCollector/internal/agent/config"
	"github.com/maffka123/metricCollector/internal/agent/models"
	"github.com/maffka123/metricCollector/internal/collector"
)

// sendDataFunc defines a function for sending data over http, it is neede for backoff.
type sendDataFunc func(context.Context, config.Config, *http.Client, []collector.MetricInterface, *zap.Logger) error

//initMetrics initializes list with all metrics of interest, send first values to the server.
func InitMetrics(ctx context.Context, cfg config.Config, client *http.Client, ch chan models.MetricList, logger *zap.Logger) {
	metricList := collector.GetAllMetrics(&cfg.Key)
	m := make([]collector.MetricInterface, len(metricList))
	for i := range metricList {
		m[i] = metricList[i]
	}

	err := simpleBackoff(ctx, sendJSONData, cfg, client, m, logger)
	if err != nil {
		ch <- models.MetricList{MetricList: nil, Err: err}
	}
	a := models.MetricList{MetricList: m, Err: nil}
	ch <- a
}

// InitPSMetrics initializes psutil metrics.
func InitPSMetrics(ctx context.Context, cfg config.Config, client *http.Client, ch chan models.MetricList, logger *zap.Logger) {

	metricList := collector.GetAllPSUtilMetrics(&cfg.Key)
	m := make([]collector.MetricInterface, len(metricList))
	for i := range metricList {
		m[i] = metricList[i]
	}

	err := simpleBackoff(ctx, sendJSONData, cfg, client, m, logger)
	if err != nil {
		ch <- models.MetricList{MetricList: nil, Err: err}
	}
	a := models.MetricList{MetricList: m, Err: nil}
	ch <- a
}

//updateMetrics updates metrics from the list.
func UpdateMetrics(ctx context.Context, cfg config.Config, cond *sync.Mutex, t <-chan time.Time, metricList []collector.MetricInterface, logger *zap.Logger) {

	// update metrics in parallel
	var wg sync.WaitGroup
	for {
		select {
		case <-t:
			// do not let sending metrics if they are now being updated
			cond.Lock()
			logger.Info("Updating all metrics")
			for _, value := range metricList {
				wg.Add(1)
				go value.Update(&wg)
			}
			wg.Wait()
			for _, value := range metricList {
				value.Print()
			}
			cond.Unlock()

		case <-ctx.Done():
			logger.Info("context canceled")
		}
	}
}

//sendJSONData sends metric in json format to the server.
func sendJSONData(ctx context.Context, cfg config.Config, client *http.Client, m []collector.MetricInterface, logger *zap.Logger) error {
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

// sendAllData iterates over metrics list and sent them to the server.
func SendAllData(ctx context.Context, cfg config.Config, cond *sync.Mutex, t <-chan time.Time, client *http.Client, metricList []collector.MetricInterface, er chan error, logger *zap.Logger) {

	// loop for allowing context cancel
	for {

		select {
		case <-t:
			// do not allow metrics to be updated while they are being sent
			cond.Lock()
			fmt.Println("Sending all metrics")
			err := simpleBackoff(ctx, sendJSONData, cfg, client, metricList, logger)
			if err != nil {
				er <- err
			}
			cond.Unlock()
		case <-ctx.Done():
			fmt.Println("context canceled")
		}
	}
}

// simpleBackoff repeats call to a function in case of an error.
func simpleBackoff(ctx context.Context, f sendDataFunc, cfg config.Config, c *http.Client, m []collector.MetricInterface, logger *zap.Logger) error {
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
			logger.Info("Backing off number", zap.String("n", fmt.Sprint(i+1)))
			time.Sleep(cfg.Delay * time.Duration(i+1))
		}
	}
	return err
}

// FanIn collects data from channels and appends is togeather.
func FanIn(outChan chan models.MetricList, inputChs ...chan models.MetricList) {
	var chInterm models.MetricList
	ch := models.MetricList{}

	for _, inputCh := range inputChs {
		chInterm = <-inputCh
		ch.MetricList = append(ch.MetricList, chInterm.MetricList...)
		if chInterm.Err != nil {
			ch.Err = chInterm.Err
			outChan <- ch
		}
	}
	outChan <- ch
}

// StartProfiling implements profiling of memory or cpu usage.
func StartProfiling(ctx chan int, file string, what string) {
	f, err := os.Create("/Users/maria/Desktop/go_intro/metricCollector/profiles/" + file)
	if err != nil {
		log.Fatal(fmt.Errorf("could not open file for profile: %v", err))
	}
	defer f.Close()

	switch {
	case what == "mem":
		<-ctx
		fmt.Println("collecting memory")
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal(fmt.Errorf("could not write heap: %v", err))
		}
		fmt.Println("collecting memory done")
		ctx <- 1

	case what == "cpu":
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(fmt.Errorf("could not write cpu: %v", err))
		}
		defer pprof.StopCPUProfile()
		<-ctx
	default:
		log.Fatal(fmt.Errorf("could not write heap: %v", what))

	}
	fmt.Println("collecting memory really done")
}
