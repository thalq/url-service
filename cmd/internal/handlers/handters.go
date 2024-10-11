package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/thalq/url-service/cmd/config"
	"github.com/thalq/url-service/cmd/internal/db_operations"
	"github.com/thalq/url-service/cmd/internal/files"
	"github.com/thalq/url-service/cmd/internal/logger"
	"github.com/thalq/url-service/cmd/internal/models"
	"github.com/thalq/url-service/cmd/internal/shortener"
	"github.com/thalq/url-service/cmd/internal/structures"
)

func PostBodyHandler(cfg config.Config, db *sql.DB, dbErr error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.Request
		var resp models.Response
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, "Не удалось прочитать тело запроса", http.StatusBadRequest)
			return
		}
		logger.Sugar.Infof("Got request: %s", buf.String())
		if err := json.Unmarshal(buf.Bytes(), &req); err != nil {
			http.Error(w, "Не удалось распарсить JSON", http.StatusBadRequest)
			return
		}
		logger.Sugar.Infof("Parsed request: %v", req)
		url := req.URL
		ifValidLink := ifValidURL(url)
		if !ifValidLink {
			http.Error(w, "Невалидный URL", http.StatusBadRequest)
			return
		}
		newLink := shortener.GenerateShortString(url)
		logger.Sugar.Infof("Generated short link: %s", newLink)

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		resp.Result = cfg.BaseURL + "/" + newLink
		response, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "Не удалось записать ответ", http.StatusInternalServerError)
			return
		}
		w.Write(response)

		if dbErr == nil {
			logger.Sugar.Infoln("Database connection established")
			URLData := structures.URLData{
				OriginalURL: url,
				ShortURL:    newLink,
			}
			if err := db_operations.ExecInsertURL(r.Context(), db, URLData); err != nil {
				logger.Sugar.Fatal(err)
			}
		} else {
			Producer, err := files.NewProducer(cfg.FileStoragePath)
			if err != nil {
				logger.Sugar.Fatal(err)
			}
			defer Producer.Close()
			var URLData = &structures.URLData{
				OriginalURL: url,
				ShortURL:    newLink,
			}
			if err := Producer.WriteEvent(URLData); err != nil {
				logger.Sugar.Fatal(err)
			}
		}
	}
}

func PostHandler(cfg config.Config, db *sql.DB, dbErr error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Не удалось прочитать тело запроса", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		bodyLink := string(body)
		ifValidLink := ifValidURL(bodyLink)
		if !ifValidLink {
			http.Error(w, "Невалидный URL", http.StatusBadRequest)
			return
		}
		newLink := shortener.GenerateShortString(bodyLink)

		logger.Sugar.Infoln("POST: Saved URL:", bodyLink, "with key:", newLink)
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(cfg.BaseURL + "/" + newLink)); err != nil {
			http.Error(w, "Не удалось записать ответ", http.StatusInternalServerError)
		}

		if dbErr == nil {
			logger.Sugar.Infoln("Database connection established")
			URLData := structures.URLData{
				OriginalURL: bodyLink,
				ShortURL:    newLink,
			}
			if err := db_operations.ExecInsertURL(r.Context(), db, URLData); err != nil {
				logger.Sugar.Fatal(err)
			}
		} else {
			Producer, err := files.NewProducer(cfg.FileStoragePath)
			if err != nil {
				logger.Sugar.Fatal(err)
			}
			defer Producer.Close()
			var URLData = &structures.URLData{
				OriginalURL: bodyLink,
				ShortURL:    newLink,
			}
			if err := Producer.WriteEvent(URLData); err != nil {
				logger.Sugar.Fatal(err)
			}
		}
	}
}

func GetHandler(cfg config.Config, db *sql.DB, dbErr error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortURL := strings.TrimPrefix(r.URL.Path, "/")
		logger.Sugar.Infoln("GET: Requested key:", shortURL)
		if dbErr == nil {
			logger.Sugar.Infoln("Database connection established")
			URLData, err := db_operations.QueryShortURL(r.Context(), db, shortURL)
			if err != nil {
				logger.Sugar.Error("ShortURL not found")
				http.Error(w, "ShortURL not found in database", http.StatusNotFound)
				return
			}
			logger.Sugar.Infoln("GET: Original URL from database:", URLData.OriginalURL)
			w.Header().Set("Location", URLData.OriginalURL)
			w.WriteHeader(http.StatusTemporaryRedirect)
			logger.Sugar.Infoln("Temporary Redirect sent for URL:", URLData.OriginalURL)
		} else {
			Consumer, err := files.NewConsumer(cfg.FileStoragePath)
			if err != nil {
				logger.Sugar.Fatal(err)
			}
			defer Consumer.Close()
			OriginalURL, err := Consumer.GetURL(shortURL)
			if err != nil {
				logger.Sugar.Error("ShortURL not found in file")
				http.Error(w, "ShortURL not found", http.StatusNotFound)
				return
			}
			logger.Sugar.Infoln("GET: Original URL from file:", OriginalURL)
			w.Header().Set("Location", OriginalURL)
			w.WriteHeader(http.StatusTemporaryRedirect)
			logger.Sugar.Infoln("Temporary Redirect sent for URL:", OriginalURL)
		}
	}
}

func GetPingHandler(cfg config.Config, err error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err != nil {
			logger.Sugar.Error("Database connection error")
			http.Error(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		logger.Sugar.Infoln("Database connection established")
		w.WriteHeader(http.StatusOK)
	}
}
