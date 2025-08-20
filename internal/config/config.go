package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddr      string `env:"SERVER_ADDRESS"`
	ReturnPrefix string `env:"BASE_URL"`
}

func ParseFlags() *Config {
	runAddr := flag.String("a", "localhost:8080", "Адрес для запуска сервера (по умолчанию: localhost:8080)")
	returnPrefix := flag.String("b", "http://localhost:8080", "Префикс для возвращаемых сокращённых URL (по умолчанию: http://localhost:8080)")

	flag.Parse()
	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		runAddr = &envRunAddr
	}
	if envReturnPrefix := os.Getenv("BASE_URL"); envReturnPrefix != "" {
		returnPrefix = &envReturnPrefix
	}

	return &Config{
		RunAddr:      *runAddr,
		ReturnPrefix: *returnPrefix,
	}
}

func NewConfig() *Config {
	return ParseFlags()
}
