package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/thalq/url-service/config"
	"github.com/thalq/url-service/internal/files"
	logger "github.com/thalq/url-service/internal/middleware"
	"github.com/thalq/url-service/internal/models"
	"github.com/thalq/url-service/internal/operations"
	"github.com/thalq/url-service/internal/shortener"
	"github.com/thalq/url-service/internal/structures"
)

func clearCookies(w http.ResponseWriter, r *http.Request) {
	// Перебираем все куки
	for _, cookie := range r.Cookies() {
		// Устанавливаем куки с истекшим временем жизни
		expiredCookie := http.Cookie{
			Name:    cookie.Name,
			Value:   "",
			Path:    "/",
			Expires: time.Unix(0, 0),
			MaxAge:  -1,
		}
		http.SetCookie(w, &expiredCookie)
		logger.Sugar.Infof("Cookie %s cleared", cookie.Name)
	}

	// Перенаправляем на главную страницу
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func PostBodyHandler(cfg config.Config, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, err := operations.GetUserId(r)
		if err != nil {
			logger.Sugar.Error("Не удалось получить ID пользователя")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var req models.Request
		var resp models.Response
		var buf bytes.Buffer
		_, err = buf.ReadFrom(r.Body)
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

		var URLData = &structures.URLData{
			CorrelationID: uuid.New().String(),
			OriginalURL:   url,
			ShortURL:      newLink,
			UserID:        userId,
		}

		resp.Result = cfg.BaseURL + "/" + newLink
		response, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "Не удалось записать ответ", http.StatusInternalServerError)
			return
		}

		if db != nil {
			if err := operations.InsertURL(r.Context(), db, URLData); err != nil {
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
		userId, err := operations.GetUserId(r)
		if err != nil {
			logger.Sugar.Error("Не удалось получить ID пользователя")
			w.WriteHeader(http.StatusUnauthorized)
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

		var URLData = &structures.URLData{
			CorrelationID: uuid.New().String(),
			OriginalURL:   bodyLink,
			ShortURL:      newLink,
			UserID:        userId,
		}
		fmt.Println(userId)
		if db != nil {
			if err := operations.InsertURL(r.Context(), db, URLData); err != nil {
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
		userId, err := operations.GetUserId(r)
		if err != nil {
			logger.Sugar.Error("Не удалось получить ID пользователя")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var batchReq []models.BatchURLRequest
		var batchResp []models.BatchURLResponse
		var URLDatas []*structures.URLData
		var buf bytes.Buffer

		_, err = buf.ReadFrom(r.Body)
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

			URLDatas = append(URLDatas, &structures.URLData{
				CorrelationID: urlReq.CorrelationID,
				OriginalURL:   urlReq.OriginalURL,
				ShortURL:      newLink,
				UserID:        userId,
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
			if err := operations.ExecInsertBatchURLs(r.Context(), db, URLDatas); err != nil {
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
		shortURL := strings.TrimPrefix(r.URL.Path, "/")
		logger.Sugar.Infoln("GET: Requested key:", shortURL)
		if db != nil {
			URLData, err := operations.GetURLData(r.Context(), db, shortURL)
			if err != nil {
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

func GetByUserHandler(cfg config.Config, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, err := operations.GetUserId(r)
		if err != nil {
			logger.Sugar.Error("Не удалось получить ID пользователя")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if db != nil {
			URLData, err := operations.GetUserURLData(r.Context(), db, userId)
			if err != nil {
				http.Error(w, "ShortURL not found in database", http.StatusNotFound)
				return
			}
			if len(URLData) == 0 {
				logger.Sugar.Info("No URLs found for user")
				w.WriteHeader(http.StatusNoContent)
				return
			}
			logger.Sugar.Infof("Get %d URLData from database", len(URLData))
			resp, err := json.Marshal(URLData)
			if err != nil {
				http.Error(w, "Не удалось записать ответ", http.StatusInternalServerError)
				return
			}
			w.Header().Set("content-type", "application/json")
			w.Write(resp)
			return
		} else {
			Consumer, err := files.NewConsumer(cfg.FileStoragePath)
			if err != nil {
				logger.Sugar.Fatal(err)
			}
			defer Consumer.Close()
			URLData, err := Consumer.GetURLsByUser(userId)
			if err != nil {
				logger.Sugar.Error("No URLs found for user")
				return
			}
			logger.Sugar.Infof("Get %d URLData from database", len(URLData))
			w.WriteHeader(http.StatusTemporaryRedirect)
			resp, err := json.Marshal(URLData)
			if err != nil {
				http.Error(w, "Не удалось записать ответ", http.StatusInternalServerError)
				return
			}
			w.Header().Set("content-type", "application/json")
			w.Write(resp)
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
