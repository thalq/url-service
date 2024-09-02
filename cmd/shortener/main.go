package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
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

func mainPage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request URL:", r.URL.Path)
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Не удалось прочитать тело запроса", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		bodyLink := string(body)

		newLink := generateShortString(bodyLink)

		URLStorage.Lock()
		URLStorage.m[newLink] = bodyLink
		URLStorage.Unlock()

		fmt.Println("POST: Saved URL:", bodyLink, "with key:", newLink)
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		fullURL := r.URL.Scheme + "://" + r.Host + r.RequestURI
		if r.URL.Scheme == "" {
			if r.TLS != nil {
				fullURL = "https://" + r.Host + r.RequestURI
			} else {
				fullURL = "http://" + r.Host + r.RequestURI
			}
		}
		if _, err := w.Write([]byte(fullURL + newLink)); err != nil {
			http.Error(w, "Не удалось записать ответ", http.StatusInternalServerError)
		}
		return
	} else if r.Method == http.MethodGet {
		url := strings.TrimPrefix(r.URL.Path, "/")

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
