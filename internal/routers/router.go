package routers

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/thalq/url-service/config"
	database "github.com/thalq/url-service/internal/dataBase"
	"github.com/thalq/url-service/internal/handlers"
	internalMiddleware "github.com/thalq/url-service/internal/middleware"
)

func NewRouter(cfg config.Config) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(internalMiddleware.WithLogging)
	r.Use(internalMiddleware.GzipMiddleware)
	r.Use(internalMiddleware.CookieMiddleware)

	db := database.DBConnect(cfg)

	r.Route("/", func(r chi.Router) {
		r.Post("/", handlers.PostHandler(cfg, db))
		r.Post("/api/shorten", handlers.PostBodyHandler(cfg, db))
		r.Post("/api/shorten/batch", handlers.PostBatchHandler(cfg, db))
		r.Get("/api/user/urls", handlers.GetByUserHandler(cfg, db))
		r.Get("/*", handlers.GetHandler(cfg, db))
		r.Get("/ping", handlers.GetPingHandler(cfg, db))
	})
	return r
}
