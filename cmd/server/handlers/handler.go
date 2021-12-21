package handlers

import (
	"fmt"
	"github.com/maffka123/metricCollector/cmd/server/storage"
	"net/http"
	"strconv"
	"strings"
)

func PostHandlerGouge(db storage.Repositories) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		}
		q := strings.Split(r.URL.String(), "/")
		val, err := strconv.ParseFloat(q[len(q)-1], 64)
		if err != nil {
			fmt.Println(err)
		}
		db.InsertGouge(q[len(q)-2], val)
		w.Header().Set("application-type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok}`))
		fmt.Printf("Got gouge: %s\n", q[len(q)-2])
	}
}

func PostHandlerCounter(db storage.Repositories) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		}
		q := strings.Split(r.URL.String(), "/")
		val, err := strconv.Atoi(q[len(q)-1])
		if err != nil {
			fmt.Println(err)
		}
		db.InsertCounter(q[len(q)-2], int64(val))
		w.Header().Set("application-type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok}`))
		fmt.Printf("Got counter: %s\n", q[len(q)-2])
	}
}
