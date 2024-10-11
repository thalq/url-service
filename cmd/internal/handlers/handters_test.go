package handlers

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/thalq/url-service/cmd/config"
	database "github.com/thalq/url-service/cmd/internal/dataBase"
	"github.com/thalq/url-service/cmd/internal/files"
	"github.com/thalq/url-service/cmd/internal/logger"
	"github.com/thalq/url-service/cmd/internal/shortener"
	"github.com/thalq/url-service/cmd/internal/structures"
	"go.uber.org/zap"
)

var testLogger *zap.Logger
var sugar *zap.SugaredLogger

func init() {
	var err error
	testLogger, err = zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Ошибка при создании логгера: %v", err)
	}
	sugar = testLogger.Sugar()
}

func TestHandlers(t *testing.T) {
	logger.Sugar = sugar

	cfg := config.ParseConfig()
	cfg.Address = "localhost:8080"
	cfg.BaseURL = "http://localhost:8080"
	logger.Sugar = sugar
	db, dbErr := database.DBConnect(cfg)

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", http.HandlerFunc(PostHandler(cfg, db, dbErr)))
		r.Post("/api/shorten", http.HandlerFunc(PostBodyHandler(cfg, db, dbErr)))
		r.Get("/*", http.HandlerFunc(GetHandler(cfg, db, dbErr)))
	})

	t.Run("POST valid URL", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader("http://example.com"))
		assert.NoError(t, err)

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "text/plain")
		assert.Contains(t, rec.Body.String(), "http://localhost:8080/")
	})

	t.Run("POST invalid URL", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader("notvalidurl"))
		assert.NoError(t, err)

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("GET non-existent URL", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/nonexist", nil)
		assert.NoError(t, err)

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("GET valid URL", func(t *testing.T) {
		shortURL := shortener.GenerateShortString("https://test.com")
		Producer, err := files.NewProducer(cfg.FileStoragePath)
		if err != nil {
			logger.Sugar.Fatal(err)
		}
		defer Producer.Close()
		var URLData = &structures.URLData{
			OriginalURL: "https://test.com",
			ShortURL:    shortURL,
		}
		if err := Producer.WriteEvent(URLData); err != nil {
			logger.Sugar.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodGet, "/"+shortURL, nil)
		assert.NoError(t, err)

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
		assert.Equal(t, "https://test.com", rec.Header().Get("Location"))
	})

	t.Run("POST valid URL with JSON body", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"https://example.com"}`))
		assert.NoError(t, err)

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
		assert.Contains(t, rec.Body.String(), "http://localhost:8080/")
	})

	t.Run("GET valid URL with JSON body", func(t *testing.T) {
		shortURL := shortener.GenerateShortString("https://test1.com")
		Producer, err := files.NewProducer(cfg.FileStoragePath)
		if err != nil {
			logger.Sugar.Fatal(err)
		}
		defer Producer.Close()
		var URLData = &structures.URLData{
			OriginalURL: "https://test1.com",
			ShortURL:    shortURL,
		}
		if err := Producer.WriteEvent(URLData); err != nil {
			logger.Sugar.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodGet, "/"+shortURL, nil)
		assert.NoError(t, err)

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
		assert.Equal(t, "https://test1.com", rec.Header().Get("Location"))
	})
}
