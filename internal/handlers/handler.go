package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"

	"github.com/maffka123/metricCollector/internal/handlers/templates"
	"github.com/maffka123/metricCollector/internal/models"
	"github.com/maffka123/metricCollector/internal/storage"
)

// MetricHandler struct to avoid repeating parsing of db and logger in every function.
type MetricHandler struct {
	db     storage.Repositories
	logger *zap.Logger
}

// metricsName type for using with html-templates.
type metricsName struct {
	NameValue string
}

// allMetricsList type allow to collects all metrics in one separated by their type.
type allMetricsList struct {
	Counter []metricsName
	Gauge   []metricsName
}

// NewMetricHandler initializes handler.
func NewMetricHandler(db storage.Repositories, logger *zap.Logger) MetricHandler {
	return MetricHandler{
		db:     db,
		logger: logger,
	}
}

// PostHandlerGouge processes POST request to add/replace value of a gouge metric.
func (mh *MetricHandler) PostHandlerGouge(dbUpdated chan time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		q := strings.Split(r.URL.String(), "/")

		val, err := strconv.ParseFloat(q[len(q)-1], 64)
		if err != nil {
			http.Error(w, "400 - Metric must be float!", http.StatusBadRequest)
			return
		}
		mh.db.InsertGouge(q[len(q)-2], val)

		w.Header().Set("application-type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok}`))
		mh.logger.Debug("Got gauge: ", zap.String("len", q[len(q)-2]))
		dbUpdated <- time.Now()
	}
}

// PostHandlerCounter processes POST request to add/replace value of a counter metric.
func (mh *MetricHandler) PostHandlerCounter(dbUpdated chan time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := strings.Split(strings.Trim(r.URL.String(), "/"), "/")

		val, err := strconv.Atoi(q[len(q)-1])
		if err != nil {
			http.Error(w, "400 - Metric must be int!", http.StatusBadRequest)
			return
		}
		mh.db.InsertCounter(q[len(q)-2], int64(val))
		w.Header().Set("application-type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok}`))
		mh.logger.Debug("Got counter: ", zap.String("len", q[len(q)-2]))
		dbUpdated <- time.Now()
	}
}

// GetHandlerValue processes GET request to return value of a specific metric.
func (mh *MetricHandler) GetHandlerValue() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		metricType := strings.ToLower(chi.URLParam(r, "type"))
		metricName := chi.URLParam(r, "name")
		if metricType == "gauge" {
			if mh.db.NameInGouge(metricName) {
				rw.Header().Set("Content-Type", "text/plain")
				rw.WriteHeader(http.StatusOK)

				payload := fmt.Sprintf("%.3f", mh.db.ValueFromGouge(metricName))
				rw.Write([]byte(payload))
			} else {
				http.Error(rw, metricName+" does not exist in Gouge db", http.StatusNotFound)
			}
		} else if metricType == "counter" {
			if mh.db.NameInCounter(metricName) {
				rw.Header().Set("Content-Type", "text/plain")
				rw.WriteHeader(http.StatusOK)

				payload := fmt.Sprintf("%d", mh.db.ValueFromCounter(metricName))
				rw.Write([]byte(payload))

			} else {
				http.Error(rw, metricName+" does not exist in Counter db", http.StatusNotFound)
			}
		} else {
			http.Error(rw, metricType+" does not exist in db", http.StatusNotFound)
		}

	}
}

// GetAllNames processes GET request to return all available metrics.
func (mh *MetricHandler) GetAllNames() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		tmpl := template.Must(template.New("MetricsList").Parse(templates.MetricTemplate))
		var aml allMetricsList
		mlc := []metricsName{}
		mlg := []metricsName{}

		rw.Header().Set("Content-Type", "text/html")
		rw.WriteHeader(http.StatusOK)
		listCounter, listGouge := mh.db.SelectAll()

		for _, v := range listCounter {
			mlc = append(mlc, metricsName{NameValue: v})

		}

		for _, v := range listGouge {
			mlg = append(mlg, metricsName{NameValue: v})

		}

		aml = allMetricsList{
			Counter: mlc,
			Gauge:   mlg,
		}

		tmpl.Execute(rw, aml)
	}
}

// PostHandlerUpdate processes POST request with json data to update particular metric.
func (mh *MetricHandler) PostHandlerUpdate(dbUpdated chan time.Time, key *string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var m models.Metrics
		err := decoder.Decode(&m)

		if err != nil {
			http.Error(w, fmt.Sprintf("400 - Metric json cannot be decoded: %s", err), http.StatusBadRequest)
			return
		}
		if m.MType == "counter" {
			mh.db.InsertCounter(m.ID, *m.Delta)
		} else {
			mh.db.InsertGouge(m.ID, *m.Value)
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
		mh.logger.Debug("Got metric: ", zap.String("name", m.ID))
		dbUpdated <- time.Now()
	}
}

// PostHandlerReturn processes POST request with json to return particular metric.
func (mh *MetricHandler) PostHandlerReturn(key *string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var m models.Metrics
		err := decoder.Decode(&m)

		if err != nil {
			http.Error(w, fmt.Sprintf("400 - Metric json cannot be decoded: %s", err), http.StatusBadRequest)
			return
		}
		if m.MType == "counter" {
			r := mh.db.ValueFromCounter(m.ID)
			m.Delta = &r
		} else {
			r := mh.db.ValueFromGouge(m.ID)
			m.Value = &r
		}

		if m.Hash != "" {
			err = m.CompareHash(*key)
			if err != nil {
				http.Error(w, "400 - Hashes do not agree", http.StatusBadRequest)
				return
			}
		}

		if *key != "" {
			m.CalcHash(*key)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		mJSON, err := json.Marshal(m)
		if err != nil {
			mh.logger.Error("JSON marshal failed: ", zap.Error(err))
		}
		w.Write([]byte(mJSON))
	}
}

// GetHandlerPing pings postgres db after GET request.
func (mh *MetricHandler) GetHandlerPing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// convert first to PGDB then to pgxpool to be able to use Ping method
		conn := mh.db.(*storage.PGDB).Conn.(*pgxpool.Pool)

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		err := conn.Ping(ctx)

		if err != nil {
			http.Error(w, "500 - Ping failed", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

	}
}

// PostHandlerUpdates updates db in batch after POST request with json data.
func (mh *MetricHandler) PostHandlerUpdates(dbUpdated chan time.Time, key *string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var ms []models.Metrics
		err := decoder.Decode(&ms)

		if err != nil {
			http.Error(w, fmt.Sprintf("400 - Metric json cannot be decoded: %s", err), http.StatusBadRequest)
			return
		}

		mh.db.BatchInsert(ms)

		w.Header().Set("application-type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok}`))

		mh.logger.Debug("got", zap.String("metrics n", fmt.Sprint(len(ms))))
		dbUpdated <- time.Now()
	}
}
