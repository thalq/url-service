package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/thalq/url-service/cmd/config"
)

// func main() {
// 	config.ParseFlags()
// 	r := chi.NewRouter()

//		r.Route("/", func(r chi.Router) {
//			r.Post("/", PostHandler)
//			r.Get("/{url}", GetHandler)
//		})
//		fmt.Println("Starting server on :8080")
//		log.Fatal(http.ListenAndServe(":8080", r))
//	}
func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	net := new(config.NetAddress)
	flag.Var(net, "a", "address and port to run server")
	flag.Parse()
	fmt.Println("Running server on", net.String())
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", PostHandler)
		r.Get("/*", GetHandler)
	})
	address := net.String()
	if address == "" {
		address = ":8080"
	}
	log.Fatal(http.ListenAndServe(address, r))
	return nil
}
