package config

import (
	"flag"
)

type Config struct {
	RunAddr      string
	ReturnPrefix string
}

func ParseFlags() *Config {
	runAddr := flag.String("a", "localhost:8080", "Адрес для запуска сервера (по умолчанию: localhost:8080)")
	returnPrefix := flag.String("b", "http://localhost:8080", "Префикс для возвращаемых сокращённых URL (по умолчанию: http://localhost:8080)")

	flag.Parse()

	return &Config{
		RunAddr:      *runAddr,
		ReturnPrefix: *returnPrefix,
	}
}

func NewConfig() *Config {
	return ParseFlags()
}
