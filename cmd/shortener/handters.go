package shortener

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/thalq/url-service/cmd/config"
)

var URLStorage = struct {
	sync.RWMutex
	m map[string]string
}{m: make(map[string]string)}

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

func PostHandler(cfg *config.Config) http.HandlerFunc {
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

		URLStorage.Lock()
		URLStorage.m[newLink] = bodyLink
		URLStorage.Unlock()

		fmt.Println("POST: Saved URL:", bodyLink, "with key:", newLink)
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(cfg.BaseURL + newLink)); err != nil {
			http.Error(w, "Не удалось записать ответ", http.StatusInternalServerError)
		}
	}
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	url := strings.TrimPrefix(r.URL.Path, "/")
	fmt.Println("GET: Requested key:", url)
	URLStorage.RLock()
	originalURL, ok := URLStorage.m[url]
	URLStorage.RUnlock()
	fmt.Println("GET: Requested key:", url)
	if ok {
		fmt.Println("GET: Found URL:", originalURL)
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
		fmt.Println("Temporary Redirect sent for URL:", originalURL)
		return
	} else {
		fmt.Println("GET: Key not found:", url)
		http.Error(w, "URL не найден", http.StatusNotFound)
		return
	}
}
