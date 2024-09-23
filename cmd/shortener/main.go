package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/thalq/url-service/cmd/config"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	cfg := config.ParseConfig()
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", PostHandler(cfg))
		r.Get("/*", GetHandler)
	})
	url := cfg.Address
	fmt.Println("Running server on", url)
	log.Fatal(http.ListenAndServe(url, r))
	return nil
}
