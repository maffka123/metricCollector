package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/maffka123/metricCollector/internal/storage"
)

func MetricRouter(db storage.Repositories) chi.Router {
	r := chi.NewRouter()

	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/update/", func(r chi.Router) {
		r.Post("/gauge/*", Conveyor(PostHandlerGouge(db), checkForPost, checkForLength))
		r.Post("/counter/*", Conveyor(PostHandlerCounter(db), checkForPost, checkForLength))
		r.Post("/*", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "501 - Metric type unknown!", http.StatusNotImplemented)
		})
		r.Post("/", Conveyor(PostHandlerUpdate(db), checkForJSON, checkForPost))

	})

	r.Route("/", func(r chi.Router) {
		r.Get("/value/{type}/{name}", GetHandlerValue(db))
		r.Get("/", GetAllNames(db))
		r.Post("/value/", Conveyor(PostHandlerReturn(db), checkForJSON, checkForPost))
	})
	return r
}
