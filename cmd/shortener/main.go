package main

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	// "github.com/go-chi/chi/middleware"
	"github.com/thalq/url-service/cmd/config"
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
		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// который будем передавать следующей функции
		ow := w

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		acceptEncoding := r.Header.Get("Accept-Encoding")
		logger.Sugar.Infof("Accept-Encoding: %s", acceptEncoding)
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		logger.Sugar.Infof("Supports gzip: %t", supportsGzip)
		if supportsGzip {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			cw := gzip.NewCompressWriter(w)
			// меняем оригинальный http.ResponseWriter на новый
			ow = cw
			// не забываем отправить клиенту все сжатые данные после завершения middleware
			defer cw.Close()
			logger.Sugar.Infof("Content-Encoding: gzip")
		}
		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		logger.Sugar.Infof("Sends gzip: %t", sendsGzip)
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := gzip.NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			r.Body = cr
			defer cr.Close()
			logger.Sugar.Infof("Content-Encoding: gzip")
		}

		// передаём управление хендлеру
		h.ServeHTTP(ow, r)
	}
}

func run() error {
	initLogger()
	cfg := config.ParseConfig()
	r := chi.NewRouter()
	// r.Use(middleware.AllowContentType("application/json", "text/html", "text/plain"))
	postHandler := logger.WithLogging(http.HandlerFunc(handlers.PostHandler(cfg)))
	postBodyHandler := logger.WithLogging(http.HandlerFunc(handlers.PostBodyHandler(cfg)))
	getHandler := logger.WithLogging(http.HandlerFunc(handlers.GetHandler))

	r.Route("/", func(r chi.Router) {
		r.Post("/", gzipMiddleware(postHandler))
		r.Post("/api/shorten", gzipMiddleware(postBodyHandler))
		r.Get("/*", getHandler)
	})
	url := cfg.Address
	sugar.Infoln("Running server on", url)
	sugar.Fatal(http.ListenAndServe(url, r))
	return nil
}
