package main

import (
	"flag"
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
	net := new(config.NetAddress)
	flag.Var(net, "a", "address and port to run server")
	flag.Var(net, "b", "address and port to run server")
	flag.Parse()
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", PostHandler)
		r.Get("/*", GetHandler)
	})
	cfg := config.ParseConfig()
	url := net.String()
	if url == "" {
		url = cfg.Address
	}
	fmt.Println("Running server on", url)
	log.Fatal(http.ListenAndServe(url, r))
	return nil
}
