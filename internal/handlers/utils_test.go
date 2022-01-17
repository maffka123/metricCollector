package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_checkForLength(t *testing.T) {
	type args struct {
		next http.Handler
	}
	type want struct {
		status int
		body   string
	}
	tests := []struct {
		name    string
		args    args
		want    want
		request string
	}{
		{name: "too short1", args: args{
			next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }),
		},
			want:    want{status: 404, body: "404 - No metric was given!\n"},
			request: "/value/"},
		{name: "too short2", args: args{
			next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }),
		},
			want:    want{status: 404, body: "404 - No metric was given!\n"},
			request: "/value/gauge/"},
		{name: "good", args: args{
			next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }),
		},
			want:    want{status: 200, body: ""},
			request: "/value/gauge/metric1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(checkForLength(tt.args.next))
			h.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()

			bodyBytes, err := io.ReadAll(result.Body)
			if err != nil {
				fmt.Println(err)
			}
			bodyString := string(bodyBytes)

			assert.Equal(t, tt.want.status, result.StatusCode)
			assert.Equal(t, tt.want.body, bodyString)
		})
	}
}

func Test_checkForPost(t *testing.T) {
	type args struct {
		next http.Handler
	}
	type want struct {
		status int
		body   string
	}
	tests := []struct {
		name    string
		args    args
		want    want
		request *http.Request
	}{
		{name: "no post", args: args{
			next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }),
		},
			want:    want{status: 405, body: "Only POST requests are allowed!\n"},
			request: httptest.NewRequest(http.MethodGet, "/update/", nil)},
		{name: "post", args: args{
			next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }),
		},
			want:    want{status: 200, body: ""},
			request: httptest.NewRequest(http.MethodPost, "/update/", nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := tt.request
			w := httptest.NewRecorder()
			h := http.HandlerFunc(checkForPost(tt.args.next))
			h.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()

			bodyBytes, err := io.ReadAll(result.Body)
			if err != nil {
				fmt.Println(err)
			}
			bodyString := string(bodyBytes)

			assert.Equal(t, tt.want.status, result.StatusCode)
			assert.Equal(t, tt.want.body, bodyString)
		})
	}
}
