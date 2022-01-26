package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/maffka123/metricCollector/internal/handlers/templates"
	"github.com/maffka123/metricCollector/internal/models"
	"github.com/maffka123/metricCollector/internal/storage"
)

type metricsList struct {
	NameValue string
}

type allMetricsList struct {
	Counter []metricsList
	Gauge   []metricsList
}

//PostHandlerGouge processes POST request to add/replace value of a gouge metric
func PostHandlerGouge(db storage.Repositories, dbUpdated chan time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		q := strings.Split(r.URL.String(), "/")

		val, err := strconv.ParseFloat(q[len(q)-1], 64)
		if err != nil {
			http.Error(w, "400 - Metric must be float!", http.StatusBadRequest)
			return
		}
		db.InsertGouge(q[len(q)-2], val)

		w.Header().Set("application-type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok}`))
		fmt.Printf("Got gauge: %s\n", q[len(q)-2])
		dbUpdated <- time.Now()
	}
}

//PostHandlerCounter processes POST request to add/replace value of a counter metric
func PostHandlerCounter(db storage.Repositories, dbUpdated chan time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := strings.Split(strings.Trim(r.URL.String(), "/"), "/")

		val, err := strconv.Atoi(q[len(q)-1])
		if err != nil {
			http.Error(w, "400 - Metric must be int!", http.StatusBadRequest)
			return
		}
		db.InsertCounter(q[len(q)-2], int64(val))
		w.Header().Set("application-type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok}`))
		fmt.Printf("Got counter: %s\n", q[len(q)-2])
		dbUpdated <- time.Now()
	}
}

//GetHandlerValue processes GET request to return value of a specific metric
func GetHandlerValue(db storage.Repositories) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		metricType := strings.ToLower(chi.URLParam(r, "type"))
		metricName := chi.URLParam(r, "name")
		if metricType == "gauge" {
			if db.NameInGouge(metricName) {
				rw.Header().Set("Content-Type", "text/plain")
				rw.WriteHeader(http.StatusOK)

				payload := fmt.Sprintf("%.3f", db.ValueFromGouge(metricName))
				rw.Write([]byte(payload))
			} else {
				http.Error(rw, metricName+" does not exist in Gouge db", http.StatusNotFound)
			}
		} else if metricType == "counter" {
			if db.NameInCounter(metricName) {
				rw.Header().Set("Content-Type", "text/plain")
				rw.WriteHeader(http.StatusOK)

				payload := fmt.Sprintf("%d", db.ValueFromCounter(metricName))
				rw.Write([]byte(payload))

			} else {
				http.Error(rw, metricName+" does not exist in Counter db", http.StatusNotFound)
			}
		} else {
			http.Error(rw, metricType+" does not exist in db", http.StatusNotFound)
		}

	}
}

//GetAllNames processes GET request to return all available metrics
func GetAllNames(db storage.Repositories) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		tmpl := template.Must(template.New("MetricsList").Parse(templates.MetricTemplate))
		var aml allMetricsList
		mlc := []metricsList{}
		mlg := []metricsList{}

		rw.Header().Set("Content-Type", "text/html")
		rw.WriteHeader(http.StatusOK)
		listCounter, listGouge := db.SelectAll()

		for _, v := range listCounter {
			mlc = append(mlc, metricsList{NameValue: v})

		}

		for _, v := range listGouge {
			mlg = append(mlg, metricsList{NameValue: v})

		}

		aml = allMetricsList{
			Counter: mlc,
			Gauge:   mlg,
		}

		tmpl.Execute(rw, aml)
	}
}

func PostHandlerUpdate(db storage.Repositories, dbUpdated chan time.Time, key *string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var m models.Metrics
		err := decoder.Decode(&m)

		if err != nil {
			http.Error(w, fmt.Sprintf("400 - Metric json cannot be decoded: %s", err), http.StatusBadRequest)
			return
		}
		if m.MType == "counter" {
			db.InsertCounter(m.ID, *m.Delta)
			fmt.Printf("Counter: %s %d\n", m.ID, *m.Delta)
		} else {
			db.InsertGouge(m.ID, *m.Value)
		}

		if key != nil && *key != "" {
			err := m.CompareHash(*key)
			if err != nil {
				http.Error(w, "400 - Hashes do not agree", http.StatusBadRequest)
				return
			}
		}

		w.Header().Set("application-type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok}`))
		fmt.Printf("Got metric: %s\n", m.ID)
		dbUpdated <- time.Now()
	}
}

func PostHandlerReturn(db storage.Repositories, key *string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var m models.Metrics
		err := decoder.Decode(&m)

		if err != nil {
			http.Error(w, fmt.Sprintf("400 - Metric json cannot be decoded: %s", err), http.StatusBadRequest)
			return
		}
		if m.MType == "counter" {
			r := db.ValueFromCounter(m.ID)
			m.Delta = &r
		} else {
			r := db.ValueFromGouge(m.ID)
			m.Value = &r
		}

		if m.Hash != "" {
			err := m.CompareHash(*key)
			if err != nil {
				http.Error(w, "400 - Hashes do not agree", http.StatusBadRequest)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		mJSON, err := json.Marshal(m)
		fmt.Println(string(mJSON))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		w.Write([]byte(mJSON))
	}
}
