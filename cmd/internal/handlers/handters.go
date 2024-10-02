package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/thalq/url-service/cmd/config"
	"github.com/thalq/url-service/cmd/internal/files"
	"github.com/thalq/url-service/cmd/internal/logger"
	"github.com/thalq/url-service/cmd/internal/models"
)

func generateShortString(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	hashBytes := h.Sum(nil)

	hexString := hex.EncodeToString(hashBytes)

	encodedString := base64.StdEncoding.EncodeToString([]byte(hexString))

	if len(encodedString) > 8 {
		return encodedString[:8]
	}
	return encodedString
}

func PostBodyHandler(cfg config.Config) http.HandlerFunc {
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
		newLink := generateShortString(url)
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
		Producer, err := files.NewProducer(cfg.FileStoragePath)
		if err != nil {
			logger.Sugar.Fatal(err)
		}
		defer Producer.Close()
		var URLData = &files.URLData{
			OriginalURL: url,
			ShortURL:    newLink,
		}
		if err := Producer.WriteEvent(URLData); err != nil {
			log.Fatal(err)
		}
	}
}

func PostHandler(cfg config.Config) http.HandlerFunc {
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
		newLink := generateShortString(bodyLink)

		logger.Sugar.Infoln("POST: Saved URL:", bodyLink, "with key:", newLink)
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(cfg.BaseURL + "/" + newLink)); err != nil {
			http.Error(w, "Не удалось записать ответ", http.StatusInternalServerError)
		}
		Producer, err := files.NewProducer(cfg.FileStoragePath)
		if err != nil {
			logger.Sugar.Fatal(err)
		}
		defer Producer.Close()
		var URLData = &files.URLData{
			OriginalURL: bodyLink,
			ShortURL:    newLink,
		}
		if err := Producer.WriteEvent(URLData); err != nil {
			log.Fatal(err)
		}
	}
}

func GetHandler(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortURL := strings.TrimPrefix(r.URL.Path, "/")
		logger.Sugar.Infoln("GET: Requested key:", shortURL)
		Consumer, err := files.NewConsumer(cfg.FileStoragePath)
		if err != nil {
			logger.Sugar.Fatal(err)
		}
		defer Consumer.Close()
		OriginalURL, err := Consumer.GetURL(shortURL)
		if err != nil {
			logger.Sugar.Error("ShortURL not found")
			http.Error(w, "ShortURL not found", http.StatusNotFound)
			return
		}
		logger.Sugar.Infoln(OriginalURL)
		w.Header().Set("Location", OriginalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
		logger.Sugar.Infoln("Temporary Redirect sent for URL:", OriginalURL)
		return
	}
}
