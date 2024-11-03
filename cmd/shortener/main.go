package main

import (
	"net/http"

	"github.com/thalq/url-service/config"
	logger "github.com/thalq/url-service/internal/middleware"
	"github.com/thalq/url-service/internal/routers"
)

func main() {
	logger.InitLogger()
	cfg := config.ParseConfig()
	r := routers.NewRouter(cfg)
	logger.Sugar.Infoln("Running server on", cfg.Address)
	logger.Sugar.Fatal(http.ListenAndServe(cfg.Address, r))
}
