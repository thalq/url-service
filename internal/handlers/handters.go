package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/thalq/url-service/config"
	"github.com/thalq/url-service/internal/ch"
	"github.com/thalq/url-service/internal/constants"
	"github.com/thalq/url-service/internal/files"
	logger "github.com/thalq/url-service/internal/middleware"
	"github.com/thalq/url-service/internal/models"
	"github.com/thalq/url-service/internal/operations"
	"github.com/thalq/url-service/internal/shortener"
)

func PostBodyHandler(cfg config.Config, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		userID, ok := ctx.Value(constants.UserIDKey).(string)
		if !ok {
			http.Error(w, "User ID not found", http.StatusUnauthorized)
			return
		}

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

		var URLData = &models.URLData{
			CorrelationID: uuid.New().String(),
			OriginalURL:   url,
			ShortURL:      newLink,
			UserID:        userID,
		}

		resp.Result = cfg.BaseURL + "/" + newLink
		response, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "Не удалось записать ответ", http.StatusInternalServerError)
			return
		}

		if db != nil {
			if err := operations.InsertURL(ctx, db, URLData); err != nil {
				logger.Sugar.Error(fmt.Sprintf("Failed to store URL: %v", err))
				w.Header().Set("content-type", "application/json")
				w.WriteHeader(http.StatusConflict)
				w.Write(response)
				return
			}
			logger.Sugar.Infoln("Data saved to database")
		} else {
			err := files.InsertDataIntoFile(cfg, URLData)
			if err != nil {
				logger.Sugar.Error(err)
			}
			logger.Sugar.Infoln("Data saved to file")
		}

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(response)
	}
}

func PostHandler(cfg config.Config, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		userID, ok := ctx.Value(constants.UserIDKey).(string)
		if !ok {
			http.Error(w, "User ID not found", http.StatusUnauthorized)
			return
		}

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

		var URLData = &models.URLData{
			CorrelationID: uuid.New().String(),
			OriginalURL:   bodyLink,
			ShortURL:      newLink,
			UserID:        userID,
		}
		if db != nil {
			if err := operations.InsertURL(ctx, db, URLData); err != nil {
				logger.Sugar.Error(fmt.Sprintf("Failed to store URL: %v", err))
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(cfg.BaseURL + "/" + URLData.ShortURL))
				return
			}
			logger.Sugar.Infoln("Data saved to database")
		} else {
			err := files.InsertDataIntoFile(cfg, URLData)
			if err != nil {
				logger.Sugar.Error(err)
			}
			logger.Sugar.Infoln("Data saved to file")
		}

		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(cfg.BaseURL + "/" + newLink)); err != nil {
			http.Error(w, "Не удалось записать ответ", http.StatusInternalServerError)
		}
	}
}

func PostBatchHandler(cfg config.Config, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 100*time.Second)
		defer cancel()

		userID, ok := ctx.Value(constants.UserIDKey).(string)
		if !ok {
			http.Error(w, "User ID not found", http.StatusUnauthorized)
			return
		}

		var batchReq []models.BatchURLRequest
		var batchResp []models.BatchURLResponse
		var URLDatas []*models.URLData
		var buf bytes.Buffer

		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, "Не удалось прочитать тело запроса", http.StatusBadRequest)
			return
		}
		logger.Sugar.Infof("Got request: %s", buf.String())
		if err := json.Unmarshal(buf.Bytes(), &batchReq); err != nil {
			http.Error(w, "Не удалось распарсить JSON", http.StatusBadRequest)
			return
		}
		logger.Sugar.Infof("Parsed request: %v", batchReq)
		for _, urlReq := range batchReq {
			if valid := ifValidURL(urlReq.OriginalURL); !valid {
				http.Error(w, "Невалидный URL", http.StatusBadRequest)
				return
			}
			newLink := shortener.GenerateShortString(urlReq.OriginalURL)
			logger.Sugar.Infof("Generated short link: %s", newLink)

			if urlReq.CorrelationID == "" {
				urlReq.CorrelationID = uuid.New().String()
			}

			URLDatas = append(URLDatas, &models.URLData{
				CorrelationID: urlReq.CorrelationID,
				OriginalURL:   urlReq.OriginalURL,
				ShortURL:      newLink,
				UserID:        userID,
			})
			batchResp = append(batchResp, models.BatchURLResponse{
				CorrelationID: urlReq.CorrelationID,
				ShortURL:      cfg.BaseURL + "/" + newLink,
			})
		}

		response, err := json.Marshal(batchResp)
		if err != nil {
			http.Error(w, "Не удалось записать ответ", http.StatusInternalServerError)
			return
		}
		if db != nil {
			logger.Sugar.Infoln("Database connection established")
			if err := operations.ExecInsertBatchURLs(ctx, db, URLDatas); err != nil {
				logger.Sugar.Error(fmt.Sprintf("Failed to store URL: %v", err))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				w.Write(response)
				return
			}
		} else {
			if err := files.InsertBatchIntoFile(cfg, URLDatas); err != nil {
				logger.Sugar.Errorf("Failed to store URL: %v", err)
			}
		}
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(response)
	}
}

func GetHandler(cfg config.Config, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		shortURL := strings.TrimPrefix(r.URL.Path, "/")
		logger.Sugar.Infoln("GET: Requested key:", shortURL)
		if db != nil {
			URLData, err := operations.GetURLData(ctx, db, shortURL)
			if err != nil {
				http.Error(w, "ShortURL not found in database", http.StatusNotFound)
				return
			}
			if URLData.DeletedFlag {
				logger.Sugar.Infoln("ShortURL is deleted")
				w.WriteHeader(http.StatusGone)
			} else {
				logger.Sugar.Infoln("GET: Original URL from database:", URLData.OriginalURL)
				w.Header().Set("Location", URLData.OriginalURL)
				w.WriteHeader(http.StatusTemporaryRedirect)
				logger.Sugar.Infoln("Temporary Redirect sent for URL:", URLData.OriginalURL)
			}
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

func GetByUserHandler(cfg config.Config, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 100*time.Second)
		defer cancel()

		userID, ok := ctx.Value(constants.UserIDKey).(string)

		fmt.Println("UserID: ", userID)
		if !ok {
			http.Error(w, "User ID not found", http.StatusUnauthorized)
			return
		}

		if db != nil {
			URLData, err := operations.GetUserURLData(ctx, db, userID)
			if err != nil {
				http.Error(w, "ShortURL not found in database", http.StatusNotFound)
				return
			}
			if len(URLData) == 0 {
				logger.Sugar.Infof("No URLs found for user %s", userID)
				w.WriteHeader(http.StatusUnauthorized) // а надо бы 204 No Content
				return
			}
			var resp []models.ShortURLData
			for _, data := range URLData {
				resp = append(resp, models.ShortURLData{
					OriginalURL: data.OriginalURL,
					ShortURL:    cfg.BaseURL + "/" + data.ShortURL,
				})
			}
			logger.Sugar.Infof("Get %d URLData from database", len(URLData))
			response, err := json.Marshal(resp)
			if err != nil {
				http.Error(w, "Не удалось записать ответ", http.StatusInternalServerError)
				return
			}
			w.Header().Set("content-type", "application/json")
			w.Write(response)
			return
		} else {
			Consumer, err := files.NewConsumer(cfg.FileStoragePath)
			if err != nil {
				logger.Sugar.Fatal(err)
			}
			defer Consumer.Close()
			URLData, err := Consumer.GetURLsByUser(userID)
			if err != nil {
				logger.Sugar.Error("No URLs found for user")
				return
			}
			logger.Sugar.Infof("Get %d URLData from database", len(URLData))

			var resp []models.ShortURLData
			for _, data := range URLData {
				resp = append(resp, models.ShortURLData{
					OriginalURL: data.OriginalURL,
					ShortURL:    cfg.BaseURL + "/" + data.ShortURL,
				})
			}
			w.WriteHeader(http.StatusTemporaryRedirect)
			response, err := json.Marshal(resp)
			if err != nil {
				http.Error(w, "Не удалось записать ответ", http.StatusInternalServerError)
				return
			}
			w.Header().Set("content-type", "application/json")
			w.Write(response)
			return
		}
	}
}

func GetPingHandler(cfg config.Config, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			logger.Sugar.Error("Database connection error")
			http.Error(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		logger.Sugar.Infoln("Database connection established")
		w.WriteHeader(http.StatusOK)
	}
}

func DeleteByList(cfg config.Config, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		userID, ok := ctx.Value(constants.UserIDKey).(string)
		if !ok {
			http.Error(w, "User ID not found", http.StatusUnauthorized)
			return
		}

		var req models.DeleteRequest
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Не удалось прочитать тело запроса", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		logger.Sugar.Infof("Got request: %s", string(body))
		if err := json.Unmarshal(body, &req.ShortURLs); err != nil {
			http.Error(w, "Не удалось распарсить JSON", http.StatusBadRequest)
			return
		}
		logger.Sugar.Infof("Parsed request: %v", req)
		w.WriteHeader(http.StatusAccepted)

		if db != nil {
			var UrlsToDelete []models.ChDelete
			for _, shortURL := range req.ShortURLs {
				UrlsToDelete = append(UrlsToDelete, models.ChDelete{
					UserID:   userID,
					ShortURL: shortURL,
				})
			}
			tx, err := db.Begin()
			if err != nil {
				logger.Sugar.Fatalf("Failed to start transaction: %v", err)
			}
			userChan := ch.Generate(UrlsToDelete...)
			results := ch.FanIn(ctx, db, userChan, tx)

			hasErrors := false
			for err := range results {
				if err != nil {
					hasErrors = true
					logger.Sugar.Infof("Error occurred: %v", err)
				}
			}

			if hasErrors {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					logger.Sugar.Infof("tx rollback error: %v", rollbackErr)
				}
				logger.Sugar.Infof("Transaction rolled back due to errors")
			} else {
				if commitErr := tx.Commit(); commitErr != nil {
					logger.Sugar.Infof("tx commit error: %v", commitErr)
				}
				logger.Sugar.Infoln("Transaction committed successfully")
			}

			logger.Sugar.Infoln("Data deleted from database")
		}
	}
}
