package handlers

import (
	"net/http"
	"strings"
)

type Middleware func(http.Handler) http.HandlerFunc

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

func checkForJson(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "This endpoint accepts only jsons", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Conveyor(h http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}
