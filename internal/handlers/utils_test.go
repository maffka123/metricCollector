package handlers

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
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

func Test_unpackGZIP(t *testing.T) {
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
		header  string
	}{
		{name: "fail", args: args{
			next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }),
		},
			want:    want{status: 405, body: "Only gzip encoding is allowed\n"},
			request: httptest.NewRequest(http.MethodPost, "/update/", nil),
			header:  "something"},
		{name: "success", args: args{
			next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				defer r.Body.Close()
				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				w.Write(bodyBytes)
			}),
		},
			want:    want{status: 200, body: `{"some": "stuff"}`},
			request: httptest.NewRequest(http.MethodPost, "/update/", gzData(`{"some": "stuff"}`)),
			header:  "gzip"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := tt.request
			request.Header.Set("Content-Encoding", tt.header)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(unpackGZIP(tt.args.next))
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

func gzData(s string) *bytes.Buffer {
	var buf bytes.Buffer
	g := gzip.NewWriter(&buf)
	if _, err := g.Write([]byte(s)); err != nil {
		fmt.Println(err)
	}
	if err := g.Close(); err != nil {
		fmt.Println(err)
	}

	return &buf
}

func Test_decodeRSA(t *testing.T) {
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
		header  string
	}{
		{name: "fail", args: args{
			next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }),
		},
			want:    want{status: 405, body: "Only gzip encoding is allowed\n"},
			request: httptest.NewRequest(http.MethodPost, "/update/", nil),
			header:  "something"},
		{name: "success", args: args{
			next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				defer r.Body.Close()
				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				w.Write(bodyBytes)
			}),
		},
			want:    want{status: 200, body: `{"some": "stuff"}`},
			request: httptest.NewRequest(http.MethodPost, "/update/", gzData2(rsaEncodedData(`{"some": "stuff"}`))),
			header:  "gzip"},
		{name: "simple success", args: args{
			next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				defer r.Body.Close()
				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				w.Write(bodyBytes)
			}),
		},
			want:    want{status: 200, body: `{"some": "stuff"}`},
			request: httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer(rsaEncodedData(`{"some": "stuff"}`))),
			header:  ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := tt.request
			if tt.header != "" {
				request.Header.Add("Content-Encoding", tt.header)
			}

			request.Header.Add("Content-Encoding", "64base")
			w := httptest.NewRecorder()
			h := http.HandlerFunc(unpackGZIP(decodeRSA(tt.args.next, getPrivKey())))
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

func rsaEncodedData(s string) []byte {
	key := getPubKey()
	encryptedBytes, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		key,
		[]byte(s),
		nil)

	if err != nil {
		log.Fatal(err)
	}
	return encryptedBytes //base64.StdEncoding.EncodeToString(encryptedBytes)
}

func getPubKey() *rsa.PublicKey {
	pub, err := ioutil.ReadFile("testdata/key.pem")
	if err != nil {
		log.Fatal(err)
	}
	pubPem, _ := pem.Decode(pub)
	pubkey, err := x509.ParsePKCS1PublicKey(pubPem.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	return pubkey
}

func getPrivKey() rsa.PrivateKey {
	pub, err := ioutil.ReadFile("testdata/key")
	if err != nil {
		log.Fatal(err)
	}
	block, _ := pem.Decode(pub)
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	return *privKey
}

func gzData2(s []byte) *bytes.Buffer {
	var buf bytes.Buffer
	g := gzip.NewWriter(&buf)
	if _, err := g.Write(s); err != nil {
		fmt.Println(err)
	}
	if err := g.Close(); err != nil {
		fmt.Println(err)
	}

	return &buf
}
