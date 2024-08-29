package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/mr-tron/base58"
)

var URLStorage = struct {
	sync.RWMutex
	m map[string]string
}{m: make(map[string]string)}

func generateShortString(s string) string {
	// 1. Создаем хэш SHA-256
	h := sha256.New()
	h.Write([]byte(s))
	hashBytes := h.Sum(nil)

	hexString := hex.EncodeToString(hashBytes)

	// 3. Используем Base58 для кодирования Hex-строки
	encodedString := base58.Encode([]byte(hexString))

	// 4. Сокращаем до длины 8 символов (или другой необходимой длины)
	if len(encodedString) > 8 {
		return encodedString[:8]
	}
	return encodedString
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request URL:", r.URL.Path)
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Не удалось прочитать тело запроса", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		bodyLink := strings.TrimSpace(string(body))

		bodyLink = strings.Trim(bodyLink, `"'`)

		newLink := generateShortString(bodyLink)

		URLStorage.Lock()
		URLStorage.m[newLink] = bodyLink
		URLStorage.Unlock()

		fmt.Println("POST: Saved URL:", bodyLink, "with key:", newLink)
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Short URL: " + newLink))
		return
	} else if r.Method == http.MethodGet {
		url := strings.TrimPrefix(r.URL.Path, "/")

		URLStorage.RLock()
		originalUrl, ok := URLStorage.m[url]
		URLStorage.RUnlock()

		fmt.Println("GET: Requested key:", url)
		if ok {
			fmt.Println("GET: Found URL:", originalUrl)
			w.Header().Set("Location", originalUrl)
			w.WriteHeader(http.StatusTemporaryRedirect)
			return
		} else {
			fmt.Println("GET: Key not found:", url)
			http.Error(w, "URL не найден", http.StatusNotFound)
			return
		}
	} else {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", mainPage)

	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
