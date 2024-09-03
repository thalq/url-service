package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
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

func main() {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/", GetHandler)
		r.Post("/", PostHandler)
	})
	fmt.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
