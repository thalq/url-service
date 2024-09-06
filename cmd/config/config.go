package config

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/caarlos0/env"
)

type Config struct {
	Address string `env:"SERVER_ADDRESS"`
	BaseURL string `env:"BASE_URL"`
}

type NetAddress struct {
	Host string
	Port int
}

func (a NetAddress) String() string {
	if a.Host == "" || a.Port == 0 {
		return ""
	}
	return a.Host + ":" + strconv.Itoa(a.Port)
}

func (a *NetAddress) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("NEED ADDRESS IN A FORM HOST:PORT")
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}
	a.Host = hp[0]
	a.Port = port
	return nil
}

func ParseConfig() *Config {
	cfg := &Config{
		Address: "localhost:8080",
		BaseURL: "http://localhost:8080",
	}
	if err := env.Parse(cfg); err != nil {
		log.Fatal(err)
	}
	return cfg
}
