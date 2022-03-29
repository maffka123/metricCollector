// Package handlers package collects routing methods for the API as well as handlers for all endpoints.
package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"crypto/rsa"
	"github.com/maffka123/metricCollector/internal/server/config"
	"github.com/maffka123/metricCollector/internal/storage"
)

// MetricRouter routes the API.
func MetricRouter(db storage.Repositories, cfg *config.Config, logger *zap.Logger) (chi.Router, chan time.Time) {
	dbUpdated := make(chan time.Time)

	r := chi.NewRouter()
	mh := NewMetricHandler(db, logger)
	rsaMW := NewRsaMW(rsa.PrivateKey(cfg.CryptoKey))

	// use inbuild middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/update/", func(r chi.Router) {
		r.Post("/gauge/*", Conveyor(mh.PostHandlerGouge(dbUpdated), checkForPost, checkForLength, unpackGZIP))
		r.Post("/counter/*", Conveyor(mh.PostHandlerCounter(dbUpdated), checkForPost, checkForLength, unpackGZIP))
		r.Post("/*", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "501 - Metric type unknown!", http.StatusNotImplemented)
		})
		r.Post("/", Conveyor(mh.PostHandlerUpdate(dbUpdated, &cfg.Key), checkForJSON, checkForPost, unpackGZIP))

	})

	r.Route("/value/", func(r chi.Router) {
		r.Get("/{type}/{name}", mh.GetHandlerValue())
		r.Post("/", Conveyor(mh.PostHandlerReturn(&cfg.Key), checkForJSON, checkForPost, packGZIP, unpackGZIP))
	})

	r.Get("/ping", mh.GetHandlerPing())
	r.Post("/updates/", Conveyor(mh.PostHandlerUpdates(dbUpdated, &cfg.Key), checkForJSON, checkForPost, rsaMW.decodeRSA, unpackGZIP))
	r.Get("/", Conveyor(mh.GetAllNames(), packGZIP))

	return r, dbUpdated
}
