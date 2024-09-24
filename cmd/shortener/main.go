package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/thalq/url-service/cmd/config"
	"github.com/thalq/url-service/cmd/internal/logger"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

var sugar *zap.SugaredLogger

func initLogger() {
	loggerInstance, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer loggerInstance.Sync()
	sugar = loggerInstance.Sugar()
	logger.sugar = *sugar
}

func run() error {
	initLogger()
	cfg := config.ParseConfig()
	r := chi.NewRouter()
	// wrappedHandler := logger.WithLogger(http.HandlerFunc(GetHandler))
	r.Route("/", func(r chi.Router) {
		r.Post("/", logger.WithLogging(PostHandler(cfg)))
		r.Get("/*", logger.WithLogging(http.HandlerFunc(GetHandler)))
	})
	url := cfg.Address
	sugar.Infoln("Running server on", url)
	sugar.Fatal(http.ListenAndServe(url, r))
	return nil
}
