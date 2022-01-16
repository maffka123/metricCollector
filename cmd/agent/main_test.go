package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_run(t *testing.T) {
	os.Setenv("BACKOFF_DELAY", "0.1s")
	tests := []struct {
		name    string
		server  bool
		wantErr bool
	}{
		{name: "no server", server: false, wantErr: true},
		//{name: "server", server: true, wantErr: false}, // running them together causes flag redefined error
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if !tt.server {
				os.Setenv("Endpoint", "127.0.0.1:8081")
			}
			srv := HttpMock("/update", http.StatusOK, struct{ Status string }{Status: "success"})
			defer srv.Close()
			err := run()
			if tt.wantErr {
				assert.Error(t, err)
			} else {

				assert.Error(t, err)

			}

		})
	}
}

type ctrl struct {
	statusCode int
	response   interface{}
}

func (c *ctrl) mockHandler(w http.ResponseWriter, r *http.Request) {
	resp := []byte{}

	rt := reflect.TypeOf(c.response)
	if rt.Kind() == reflect.String {
		resp = []byte(c.response.(string))
	} else if rt.Kind() == reflect.Struct || rt.Kind() == reflect.Ptr {
		resp, _ = json.Marshal(c.response)
	} else {
		resp = []byte("{}")
	}

	w.WriteHeader(c.statusCode)
	w.Write(resp)
}

func HttpMock(pattern string, statusCode int, response interface{}) *httptest.Server {
	c := &ctrl{statusCode, response}

	handler := http.NewServeMux()
	handler.HandleFunc(pattern, c.mockHandler)

	return httptest.NewServer(handler)
}
