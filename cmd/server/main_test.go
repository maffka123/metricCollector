// server to collect metrics coming to its endpoints.
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func Example() {
	go main()

	client := &http.Client{}

	// Send gauge as a part of link
	url := "http://localhost:8080/update/gauge/Alloc/0.5"
	request, _ := http.NewRequest(http.MethodPost, url, nil)
	response, _ := client.Do(request)
	defer response.Body.Close()

	fmt.Println(response.Status)

	// Send counter as a part of link
	url = "http://localhost:8080/update/counter/RandomValue/4"
	request, _ = http.NewRequest(http.MethodPost, url, nil)
	response, _ = client.Do(request)
	defer response.Body.Close()

	fmt.Println(response.Status)

	// Get gauge value
	url = "http://localhost:8080/value/gauge/Alloc"
	request, _ = http.NewRequest(http.MethodGet, url, nil)
	response, _ = client.Do(request)
	defer response.Body.Close()

	responseData, _ := ioutil.ReadAll(response.Body)

	fmt.Println(string(responseData))

	// Get counter value
	url = "http://localhost:8080/value/counter/RandomValue"
	request, _ = http.NewRequest(http.MethodGet, url, nil)
	response, _ = client.Do(request)
	defer response.Body.Close()

	responseData, _ = ioutil.ReadAll(response.Body)

	fmt.Println(string(responseData))

	// Output: 200 OK
	// 200 OK
	// 0.500
	// 4
}
