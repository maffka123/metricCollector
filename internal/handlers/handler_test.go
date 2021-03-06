package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/caarlos0/env/v6"
	"github.com/driftprogramming/pgxpoolmock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	globalConf "github.com/maffka123/metricCollector/internal/config"
	"github.com/maffka123/metricCollector/internal/models"
	"github.com/maffka123/metricCollector/internal/server/config"
	"github.com/maffka123/metricCollector/internal/storage"
)

var logger *zap.Logger = globalConf.InitLogger(true)

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
	db := storage.Connect(cfg, logger)

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
				valueInDB:       0.5,
			},
			request: "/update/gauge/RandomValue/0.5",
		},
		{
			name: "gauge_handler_replace",
			args: args{db: db},
			want: want{
				applicationType: "text/plain",
				statusCode:      200,
				valueInDB:       1.3,
			},
			request: "/update/gauge/RandomValue/1.3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, dbUpdated := MetricRouter(db, cfg, logger)

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
	db := storage.Connect(cfg, logger)
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
			r, dbUpdated := MetricRouter(db, cfg, logger)
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
	db := storage.Connect(cfg, logger)
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
			r, dbUpdated := MetricRouter(db, cfg, logger)
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
	db := storage.Connect(cfg, logger)
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
				html:        "<html>\n    <head>\n    <title>(/^???^)/</title>\n    </head>\n    <body>\n        <h1>Counter</h1>>\n    \n            <li>[PollCount]: [3]\n</li>\n    \n\n    <h1>Gauge</h1>>\n    \n    <li>[Alloc]: [1.500]\n</li>\n\n\n    </body>\n</html>",
			},
			request: "/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, dbUpdated := MetricRouter(db, cfg, logger)
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
	cfg.Key = "test"
	db := storage.Connect(cfg, logger)
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
			request: request{request: "/update/", body: models.Metrics{ID: "Alloc", MType: "gauge", Value: &f, Hash: "bd4208a757a7c5e94a4ce2975530aaddadf889c8ee627798e57e89eb066d6c3d"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r, dbUpdated := MetricRouter(db, cfg, logger)
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
	db := storage.Connect(cfg, logger)
	cfg.Key = "test"
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
				json:            models.Metrics{ID: "Alloc", MType: "gauge", Value: &f, Hash: "bd4208a757a7c5e94a4ce2975530aaddadf889c8ee627798e57e89eb066d6c3d"},
			},
			request: request{request: "/value/", body: models.Metrics{ID: "Alloc", MType: "gauge", Hash: "bd4208a757a7c5e94a4ce2975530aaddadf889c8ee627798e57e89eb066d6c3d"}},
		},
		{
			name: "gauge2",
			args: args{db: db},
			want: want{
				applicationType: "application/json",
				statusCode:      200,
				json:            models.Metrics{ID: "Alloc", MType: "gauge", Value: &f, Hash: "bd4208a757a7c5e94a4ce2975530aaddadf889c8ee627798e57e89eb066d6c3d"},
			},
			request: request{request: "/value/", body: models.Metrics{ID: "Alloc", MType: "gauge"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.InsertGouge("Alloc", float64(1.5))
			r, dbUpdated := MetricRouter(db, cfg, logger)
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
			fmt.Println(string(body))
			assert.Equal(t, want, bodyBytes)
		})
	}
}

func TestPostHandlerUpdateWithPostgres(t *testing.T) {
	f := float64(1.5)
	cfg := prepConf()

	cfg.Key = "test"
	ctrl := gomock.NewController(t)
	mockdb := pgxpoolmock.NewMockPgxPool(ctrl)

	mockdb.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte("test"), nil)

	db := storage.ConnectPG(context.Background(), cfg, logger)
	db.Conn = mockdb

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
		want    want
		request request
	}{
		{
			name: "gauge",
			want: want{
				applicationType: "text/plain",
				statusCode:      200,
				valueInDB:       1.5,
			},
			request: request{request: "/update/", body: models.Metrics{ID: "Alloc", MType: "gauge", Value: &f, Hash: "bd4208a757a7c5e94a4ce2975530aaddadf889c8ee627798e57e89eb066d6c3d"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r, dbUpdated := MetricRouter(db, cfg, logger)
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
		})
	}
}

// seems there is a bug for returning single value
/*func TestPostHandlerReturnWithPG(t *testing.T) {
	f := float64(1.5)
	cfg := prepConf()
	cfg.Key = "test"
	ctrl := gomock.NewController(t)
	mockdb := pgxpoolmock.NewMockPgxPool(ctrl)

	pgxRows := pgxpoolmock.NewRows([]string{"value"}).AddRow(1.5).ToPgxRows()

	mockdb.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any()).Return(pgxRows)

	db := storage.ConnectPG(context.Background(), cfg, logger)
	db.Conn = mockdb
	cfg.Key = "test"

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
		want    want
		request request
	}{
		{
			name: "gauge",
			want: want{
				applicationType: "application/json",
				statusCode:      200,
				json:            models.Metrics{ID: "Alloc", MType: "gauge", Value: &f, Hash: "bd4208a757a7c5e94a4ce2975530aaddadf889c8ee627798e57e89eb066d6c3d"},
			},
			request: request{request: "/value/", body: models.Metrics{ID: "Alloc", MType: "gauge", Hash: "bd4208a757a7c5e94a4ce2975530aaddadf889c8ee627798e57e89eb066d6c3d"}},
		},
		{
			name: "gauge2",
			want: want{
				applicationType: "application/json",
				statusCode:      200,
				json:            models.Metrics{ID: "Alloc", MType: "gauge", Value: &f, Hash: "bd4208a757a7c5e94a4ce2975530aaddadf889c8ee627798e57e89eb066d6c3d"},
			},
			request: request{request: "/value/", body: models.Metrics{ID: "Alloc", MType: "gauge"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r, dbUpdated := MetricRouter(db, &cfg.Key, logger)
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
			fmt.Println(string(body))
			assert.Equal(t, want, bodyBytes)
		})
	}
}*/
