package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/maffka123/metricCollector/cmd/server/storage"
)

func TestPostHandlerGouge(t *testing.T) {
	db := storage.NewInMemoryDB()
	type args struct {
		db storage.Repositories
	}
	type want struct {
		applicationType string
		statusCode      int
		valueInDb       float64
	}
	tests := []struct {
		name    string
		args    args
		want    want
		request string
	}{
		{
			name: "gouge_handler_test1",
			args: args{db: db},
			want: want{
				applicationType: "text/plain",
				statusCode:      200,
				valueInDb:       1,
			},
			request: "/update/gouge/RandomValue/1",
		},
		{
			name: "gouge_handler_replace",
			args: args{db: db},
			want: want{
				applicationType: "text/plain",
				statusCode:      200,
				valueInDb:       2,
			},
			request: "/update/gouge/RandomValue/2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(PostHandlerGouge(tt.args.db))
			h.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.applicationType, result.Header.Get("Application-Type"))
			q := strings.Split(request.URL.String(), "/")
			assert.Equal(t, tt.want.valueInDb, db.Gouge[q[len(q)-2]])
		})
	}
}

func TestPostHandlerCounter(t *testing.T) {
	db := storage.NewInMemoryDB()
	type args struct {
		db storage.Repositories
	}
	type want struct {
		applicationType string
		statusCode      int
		value_in_db     int64
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
				value_in_db:     1,
			},
			request: "/update/count/PollCount/1",
		},
		{
			name: "count_handler_increment",
			args: args{db: db},
			want: want{
				applicationType: "text/plain",
				statusCode:      200,
				value_in_db:     3,
			},
			request: "/update/count/PollCount/2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(PostHandlerCounter(tt.args.db))
			h.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.applicationType, result.Header.Get("Application-Type"))
			q := strings.Split(request.URL.String(), "/")
			assert.Equal(t, tt.want.value_in_db, db.Counter[q[len(q)-2]])
		})
	}
}
