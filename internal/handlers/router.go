package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/maffka123/metricCollector/internal/storage"
)

func MetricRouter(db storage.Repositories, key *string, logger *zap.Logger) (chi.Router, chan time.Time) {
	dbUpdated := make(chan time.Time)

	r := chi.NewRouter()
	mh := NewMetricHandler(db, logger)

	// зададим встроенные middleware, чтобы улучшить стабильность приложения
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
		r.Post("/", Conveyor(mh.PostHandlerUpdate(dbUpdated, key), checkForJSON, checkForPost, unpackGZIP))

	})

	r.Route("/value/", func(r chi.Router) {
		r.Get("/{type}/{name}", mh.GetHandlerValue())
		r.Post("/", Conveyor(mh.PostHandlerReturn(key), checkForJSON, checkForPost, packGZIP, unpackGZIP))
	})

	r.Get("/ping", mh.GetHandlerPing())
	r.Post("/updates/", Conveyor(mh.PostHandlerUpdates(dbUpdated, key), checkForJSON, checkForPost, unpackGZIP))
	r.Get("/", Conveyor(mh.GetAllNames(), packGZIP))

	return r, dbUpdated
}
