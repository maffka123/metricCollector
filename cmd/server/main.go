package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

var db = map[string]string{}

func main() {
	http.HandleFunc("/update/", postHandler)
	fmt.Println("Start serving on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
	}
	q := strings.Split(r.URL.String(), "/")
	db[q[len(q)-2]] = q[len(q)-1]
	fmt.Println(db)
	fmt.Fprintf(w, "Got %s!", q[len(q)-2])
}
