package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/maffka123/metricCollector/internal/storage"
	"go.uber.org/zap"
)

func MetricRouter(db storage.Repositories, key *string, logger *zap.Logger) (chi.Router, chan time.Time) {
	dbUpdated := make(chan time.Time)

	r := chi.NewRouter()

	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/update/", func(r chi.Router) {
		r.Post("/gauge/*", Conveyor(PostHandlerGouge(db, dbUpdated, logger), checkForPost, checkForLength, unpackGZIP))
		r.Post("/counter/*", Conveyor(PostHandlerCounter(db, dbUpdated, logger), checkForPost, checkForLength, unpackGZIP))
		r.Post("/*", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "501 - Metric type unknown!", http.StatusNotImplemented)
		})
		r.Post("/", Conveyor(PostHandlerUpdate(db, dbUpdated, key, logger), checkForJSON, checkForPost, unpackGZIP))

	})

	r.Route("/value/", func(r chi.Router) {
		r.Get("/{type}/{name}", GetHandlerValue(db))
		r.Post("/", Conveyor(PostHandlerReturn(db, key, logger), checkForJSON, checkForPost, packGZIP, unpackGZIP))
	})

	r.Get("/ping", GetHandlerPing(db))
	r.Post("/updates/", Conveyor(PostHandlerUpdates(db, dbUpdated, key, logger), checkForJSON, checkForPost, unpackGZIP))
	r.Get("/", Conveyor(GetAllNames(db), packGZIP))

	return r, dbUpdated
}
