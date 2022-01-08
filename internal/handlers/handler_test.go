package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/caarlos0/env/v6"
	"github.com/maffka123/metricCollector/internal/models"
	"github.com/maffka123/metricCollector/internal/server/config"
	"github.com/maffka123/metricCollector/internal/storage"
	"github.com/stretchr/testify/assert"
)

func prepConf() *config.Config {
	var cfg config.Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	cfg.Restore = false
	return &cfg
}

func TestPostHandlerGouge(t *testing.T) {
	cfg := prepConf()
	db := storage.Connect(cfg)
	type args struct {
		db storage.Repositories
	}
	type want struct {
		applicationType string
		statusCode      int
		valueInDB       float64
	}
	tests := []struct {
		name    string
		args    args
		want    want
		request string
	}{
		{
			name: "gauge_handler_test1",
			args: args{db: db},
			want: want{
				applicationType: "text/plain",
				statusCode:      200,
				valueInDB:       1,
			},
			request: "/update/gauge/RandomValue/1",
		},
		{
			name: "gauge_handler_replace",
			args: args{db: db},
			want: want{
				applicationType: "text/plain",
				statusCode:      200,
				valueInDB:       2,
			},
			request: "/update/gauge/RandomValue/2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, dbUpdated := MetricRouter(db)

			go func() { <-dbUpdated }()

			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.applicationType, result.Header.Get("Application-Type"))
			q := strings.Split(request.URL.String(), "/")
			assert.Equal(t, tt.want.valueInDB, db.Gouge[q[len(q)-2]])
		})
	}
}

func TestPostHandlerCounter(t *testing.T) {
	cfg := prepConf()
	db := storage.Connect(cfg)
	type args struct {
		db storage.Repositories
	}
	type want struct {
		applicationType string
		statusCode      int
		valueInDB       int64
	}
	tests := []struct {
		name    string
		args    args
		want    want
		request string
	}{
		{
			name: "count_handler_test1",
			args: args{db: db},
			want: want{
				applicationType: "text/plain",
				statusCode:      200,
				valueInDB:       1,
			},
			request: "/update/counter/PollCount/1",
		},
		{
			name: "count_handler_increment",
			args: args{db: db},
			want: want{
				applicationType: "text/plain",
				statusCode:      200,
				valueInDB:       3,
			},
			request: "/update/counter/PollCount/2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, dbUpdated := MetricRouter(db)
			go func() { <-dbUpdated }()

			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.applicationType, result.Header.Get("Application-Type"))
			q := strings.Split(request.URL.String(), "/")
			assert.Equal(t, tt.want.valueInDB, db.Counter[q[len(q)-2]])
		})
	}
}

func TestGetHandlerValue(t *testing.T) {
	cfg := prepConf()
	db := storage.Connect(cfg)
	db.InsertCounter("PollCount", 3)
	db.InsertGouge("Alloc", 1)
	type args struct {
		db storage.Repositories
	}
	type want struct {
		contentType string
		statusCode  int
		valueInDB   string
	}
	tests := []struct {
		name    string
		args    args
		want    want
		request string
	}{
		{
			name: "count_handler_test1",
			args: args{db: db},
			want: want{
				contentType: "text/plain",
				statusCode:  200,
				valueInDB:   "1.000",
			},
			request: "/value/gauge/Alloc",
		},
		{
			name: "count_handler_increment",
			args: args{db: db},
			want: want{
				contentType: "text/plain",
				statusCode:  200,
				valueInDB:   "3",
			},
			request: "/value/counter/PollCount",
		},
		{
			name: "count_unknown",
			args: args{db: db},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  404,
				valueInDB:   "SomeCount does not exist in Counter db\n",
			},
			request: "/value/counter/SomeCount",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, dbUpdated := MetricRouter(db)
			go func() { <-dbUpdated }()

			request := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			fmt.Println(w.Body.String())

			assert.Equal(t, tt.want.valueInDB, w.Body.String())
		})
	}
}

func TestGetAllNames(t *testing.T) {
	cfg := prepConf()
	db := storage.Connect(cfg)
	db.InsertCounter("PollCount", 3)
	db.InsertGouge("Alloc", 1.5)
	type args struct {
		db storage.Repositories
	}
	type want struct {
		contentType string
		statusCode  int
		html        string
	}
	tests := []struct {
		name    string
		args    args
		want    want
		request string
	}{
		{
			name: "allmetrics_test1",
			args: args{db: db},
			want: want{
				contentType: "text/html",
				statusCode:  200,
				html:        "<html>\n    <head>\n    <title>(/^▽^)/</title>\n    </head>\n    <body>\n        <h1>Counter</h1>>\n    \n            <li>[PollCount]: [3]\n</li>\n    \n\n    <h1>Gauge</h1>>\n    \n    <li>[Alloc]: [1.500]\n</li>\n\n\n    </body>\n</html>",
			},
			request: "/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, dbUpdated := MetricRouter(db)
			go func() { <-dbUpdated }()

			request := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			fmt.Println(w.Body.String())

			assert.Equal(t, tt.want.html, w.Body.String())
		})
	}
}

func TestPostHandlerUpdate(t *testing.T) {
	f := float64(1.5)
	cfg := prepConf()
	db := storage.Connect(cfg)
	type args struct {
		db storage.Repositories
	}
	type want struct {
		applicationType string
		statusCode      int
		valueInDB       float64
	}
	type request struct {
		request string
		body    models.Metrics
	}
	tests := []struct {
		name    string
		args    args
		want    want
		request request
	}{
		{
			name: "gauge",
			args: args{db: db},
			want: want{
				applicationType: "text/plain",
				statusCode:      200,
				valueInDB:       1.5,
			},
			request: request{request: "/update/", body: models.Metrics{ID: "Alloc", MType: "gauge", Value: &f}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r, dbUpdated := MetricRouter(db)
			go func() { <-dbUpdated }()
			body, _ := json.Marshal(tt.request.body)
			request := httptest.NewRequest(http.MethodPost, tt.request.request, bytes.NewBuffer(body))
			request.Header.Add("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.applicationType, result.Header.Get("Application-Type"))
			assert.Equal(t, tt.want.valueInDB, db.Gouge[tt.request.body.ID])
		})
	}
}

func TestPostHandlerReturn(t *testing.T) {
	f := float64(1.5)
	cfg := prepConf()
	db := storage.Connect(cfg)
	type args struct {
		db storage.Repositories
	}
	type want struct {
		applicationType string
		statusCode      int
		json            models.Metrics
	}
	type request struct {
		request string
		body    models.Metrics
	}
	tests := []struct {
		name    string
		args    args
		want    want
		request request
	}{
		{
			name: "gauge",
			args: args{db: db},
			want: want{
				applicationType: "application/json",
				statusCode:      200,
				json:            models.Metrics{ID: "Alloc", MType: "gauge", Value: &f},
			},
			request: request{request: "/value/", body: models.Metrics{ID: "Alloc", MType: "gauge"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.InsertGouge("Alloc", float64(1.5))
			r, dbUpdated := MetricRouter(db)
			go func() { <-dbUpdated }()
			body, _ := json.Marshal(tt.request.body)
			request := httptest.NewRequest(http.MethodPost, tt.request.request, bytes.NewBuffer(body))
			request.Header.Add("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.applicationType, result.Header.Get("Content-Type"))
			want, _ := json.Marshal(tt.want.json)
			bodyBytes, _ := io.ReadAll(result.Body)
			assert.Equal(t, want, bodyBytes)
		})
	}
}
