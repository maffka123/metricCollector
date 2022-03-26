package handlers

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	//"encoding/base64"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type Middleware func(http.Handler) http.HandlerFunc
type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func checkForLength(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := strings.Split(strings.Trim(r.URL.String(), "/"), "/")
		if len(q) < 3 {
			http.Error(w, "404 - No metric was given!", http.StatusNotFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func checkForPost(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func checkForJSON(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "This endpoint accepts only jsons", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func unpackGZIP(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, b := range r.Header.Values("Content-Encoding") {
			if b != "gzip" && b != "64base" {
				http.Error(w, "Only gzip encoding is allowed", http.StatusMethodNotAllowed)
				return
			}
		}
		if len(r.Header.Values("Content-Encoding")) == 0 ||
			(len(r.Header.Values("Content-Encoding")) == 1 && r.Header.Get("Content-Encoding") == "64base") {
			next.ServeHTTP(w, r)
			return
		}

		rw, err := gzip.NewReader(r.Body)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		r.Body = rw
		next.ServeHTTP(w, r)
	})
}

func decodeRSA(next http.Handler, key rsa.PrivateKey) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, b := range r.Header.Values("Content-Encoding") {
			if b != "gzip" && b != "64base" {
				http.Error(w, "Only 64base encoding is allowed", http.StatusMethodNotAllowed)
				return
			}
		}
		if len(r.Header.Values("Content-Encoding")) == 0 ||
			(len(r.Header.Values("Content-Encoding")) == 1 && r.Header.Get("Content-Encoding") == "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		hash := sha256.New()
		drb, err := rsa.DecryptOAEP(hash, rand.Reader, &key, body, nil)
		if err != nil {
			log.Fatal(err)
		}
		r.Body = ioutil.NopCloser(bytes.NewBufferString(string(drb)))
		next.ServeHTTP(w, r)
	})
}

func packGZIP(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

func Conveyor(h http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}
