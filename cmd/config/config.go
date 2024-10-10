package config

import (
	"flag"
	"os"

	"github.com/thalq/url-service/cmd/internal/logger"
)

type Config struct {
	Address         string `env:"SERVER_ADDRESS" json:"address"`
	BaseURL         string `env:"BASE_URL" json:"base_url"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DatabaseDNS     string `env:"DATABASE_DSN" json:"database_dns"`
}

func getEnv(value string, defaultValue string) string {
	if value, exists := os.LookupEnv(value); exists {
		return value
	}
	return defaultValue
}

func ParseConfig() Config {
	defaultAddress := "localhost:8080"
	defaultBaseURL := "http://localhost:8080"
	currDir, err := os.Getwd()
	if err != nil {
		logger.Sugar.Fatalf("Ошибка при получении текущего каталога: %v", err)
	}
	defaultFileStoragePath := currDir + "/url_data.log"
	defaultDatabaseDNS := "postgres://postgres:postgres@localhost/postgres?sslmode=disable"

	envAddress := getEnv("SERVER_ADDRESS", defaultAddress)
	envBaseURL := getEnv("BASE_URL", defaultBaseURL)
	envFileStoragePath := getEnv("FILE_STORAGE_PATH", defaultFileStoragePath)
	envDatabaseDNS := getEnv("DATABASE_DSN", defaultDatabaseDNS)

	logger.Sugar.Infof("Address: %s; BaseURL: %s; FileStoragePath: %s", envAddress, envBaseURL, envFileStoragePath)

	address := flag.String("a", envAddress, "address to run server")
	baseURL := flag.String("b", envBaseURL, "port to run server")
	fileStoragePath := flag.String("f", envFileStoragePath, "path to file storage")
	databaseDNS := flag.String("d", envDatabaseDNS, "database DSN")

	flag.Parse()
	return Config{
		Address:         *address,
		BaseURL:         *baseURL,
		FileStoragePath: *fileStoragePath,
		DatabaseDNS:     *databaseDNS,
	}
}
