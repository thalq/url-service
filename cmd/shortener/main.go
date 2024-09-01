package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
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
		PostHandler(w, r)
		return
	} else if r.Method == http.MethodGet {
		GetHandler(w, r)
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
