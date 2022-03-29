package server

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"

	"encoding/json"

	"github.com/maffka123/metricCollector/internal/handlers"
	"github.com/maffka123/metricCollector/internal/models"
	"github.com/maffka123/metricCollector/internal/server/config"
	"github.com/maffka123/metricCollector/internal/storage"
	"go.uber.org/zap"
)

func Example() {
	// Init config and logger
	cfg := config.Config{Endpoint: "localhost:8086", Key: "testkey", Restore: false}
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("Could not initialize logger: " + err.Error())
	}

	// Init in-memory db and flush dbupdated channel, otherwise it blocks everything
	db := storage.Connect(&cfg, logger)
	r, dbUpdated := handlers.MetricRouter(db, &cfg, logger)
	go func() {
		for {
			<-dbUpdated
		}
	}()

	// Start server
	srv := httptest.NewUnstartedServer(r)
	srv.Listener.Close()
	srv.Listener, err = net.Listen("tcp", "localhost:8086")
	if err != nil {
		log.Fatal("Could not start test server: " + err.Error())
	}
	srv.Start()
	defer srv.Close()

	// does not work in git-tests, start of goroutine is undefined
	/*srv := NewServer(cfg.Endpoint, r)
	go func() { log.Fatal(srv.ListenAndServe()) }()*/

	client := srv.Client()

	// Send gauge as a part of link
	url := "http://localhost:8086/update/gauge/Alloc/0.5"
	request, _ := http.NewRequest(http.MethodPost, url, nil)
	response, err := client.Do(request)
	if err != nil {
		log.Fatal("Send gauge as a part of link: " + err.Error())
	}
	defer response.Body.Close()

	fmt.Println(response.Status)

	// Send counter as a part of link
	url = "http://localhost:8086/update/counter/RandomValue/4"
	request, _ = http.NewRequest(http.MethodPost, url, nil)
	response, err = client.Do(request)
	if err != nil {
		log.Fatal("Send counter as a part of link: " + err.Error())
	}
	defer response.Body.Close()

	fmt.Println(response.Status)

	// Get gauge value
	url = "http://localhost:8086/value/gauge/Alloc"
	request, _ = http.NewRequest(http.MethodGet, url, nil)
	response, err = client.Do(request)
	if err != nil {
		log.Fatal("Get gauge value: " + err.Error())
	}
	defer response.Body.Close()

	responseData, _ := ioutil.ReadAll(response.Body)

	fmt.Println(string(responseData))

	// Get counter value
	url = "http://localhost:8086/value/counter/RandomValue"
	request, _ = http.NewRequest(http.MethodGet, url, nil)
	response, err = client.Do(request)
	if err != nil {
		log.Fatal("Get counter value: " + err.Error())
	}
	defer response.Body.Close()

	responseData, _ = ioutil.ReadAll(response.Body)

	fmt.Println(string(responseData))

	// Get unknown value
	url = "http://localhost:8086/value/counter/Unknown"
	request, _ = http.NewRequest(http.MethodGet, url, nil)
	response, err = client.Do(request)
	if err != nil {
		log.Fatal("Get counter value: " + err.Error())
	}
	defer response.Body.Close()

	responseData, _ = ioutil.ReadAll(response.Body)

	fmt.Println(string(responseData))

	// Get as json
	url = "http://localhost:8086/value/"
	m := models.Metrics{MType: "counter", ID: "RandomValue"}
	data, err := json.Marshal(m)
	if err != nil {
		log.Fatal("Could not marshall metric: " + err.Error())
	}
	request, _ = http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	request.Header.Add("Content-Type", "application/json")
	response, err = client.Do(request)
	if err != nil {
		log.Fatal("Get counter value: " + err.Error())
	}
	defer response.Body.Close()

	responseData, _ = ioutil.ReadAll(response.Body)

	fmt.Println(string(responseData))

	// Output: 200 OK
	// 200 OK
	// 0.500
	// 4
	// Unknown does not exist in Counter db
	//
	// {"id":"RandomValue","type":"counter","delta":4,"hash":"8def3c216b2fb10be4531880af9f3958808fdb97d2ab79040c7ea661e04459a2"}
}
