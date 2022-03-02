// server to collect metrics coming to its endpoints.
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func Example() {
	go main()

	client := &http.Client{}

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

	// Output: 200 OK
	// 200 OK
	// 0.500
	// 4
}
