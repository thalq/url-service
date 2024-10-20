package routers

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/thalq/url-service/config"
	database "github.com/thalq/url-service/internal/dataBase"
	"github.com/thalq/url-service/internal/handlers"
	"github.com/thalq/url-service/internal/logger"
	internalMiddleware "github.com/thalq/url-service/internal/middleware"
)

func NewRouter(cfg config.Config) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(logger.WithLogging)
	r.Use(internalMiddleware.GzipMiddleware)

	db := database.DBConnect(cfg)

	r.Route("/", func(r chi.Router) {
		r.Post("/", handlers.PostHandler(cfg, db))
		r.Post("/api/shorten", handlers.PostBodyHandler(cfg, db))
		r.Post("/api/shorten/batch", handlers.PostBatchHandler(cfg, db))
		r.Get("/*", handlers.GetHandler(cfg, db))
		r.Get("/ping", handlers.GetPingHandler(cfg, db))
	})
	return r
}
