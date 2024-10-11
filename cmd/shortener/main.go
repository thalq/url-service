package main

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/thalq/url-service/cmd/config"
	"github.com/thalq/url-service/cmd/internal/dataBase"
	"github.com/thalq/url-service/cmd/internal/gzip"
	"github.com/thalq/url-service/cmd/internal/handlers"
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
	logger.Sugar = sugar
}

func gzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		logger.Sugar.Infof("Accept-Encoding: %s", acceptEncoding)
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		logger.Sugar.Infof("Supports gzip: %t", supportsGzip)
		if supportsGzip {
			cw := gzip.NewCompressWriter(w)
			ow = cw
			defer cw.Close()
			logger.Sugar.Infof("Content-Encoding: gzip")
		}
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		logger.Sugar.Infof("Sends gzip: %t", sendsGzip)
		if sendsGzip {
			cr, err := gzip.NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
			logger.Sugar.Infof("Content-Encoding: gzip")
		}

		h.ServeHTTP(ow, r)
	}
}

func run() error {
	initLogger()
	cfg := config.ParseConfig()
	r := chi.NewRouter()
	db, err := dataBase.DBConnect(cfg)
	postHandler := logger.WithLogging(http.HandlerFunc(handlers.PostHandler(cfg, db, err)))
	postBodyHandler := logger.WithLogging(http.HandlerFunc(handlers.PostBodyHandler(cfg, db, err)))
	getHandler := logger.WithLogging(http.HandlerFunc(handlers.GetHandler(cfg, db, err)))
	getPingHandler := logger.WithLogging(http.HandlerFunc(handlers.GetPingHandler(cfg, err)))

	r.Route("/", func(r chi.Router) {
		r.Post("/", gzipMiddleware(postHandler))
		r.Post("/api/shorten", gzipMiddleware(postBodyHandler))
		r.Get("/*", getHandler)
		r.Get("/ping", getPingHandler)
	})
	url := cfg.Address
	sugar.Infoln("Running server on", url)
	sugar.Fatal(http.ListenAndServe(url, r))
	return nil
}
