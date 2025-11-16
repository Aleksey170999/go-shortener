package config

import (
	"flag"
	"os"

	"github.com/Aleksey170999/go-shortener/internal/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	RunAddr         string `env:"SERVER_ADDRESS"`
	ReturnPrefix    string `env:"BASE_URL"`
	Logger          zap.Logger
	StorageFilePath string
	DatabaseDSN     string
	AuditURL        string
	AuditFile       string
}

func ParseFlags() *Config {
	runAddr := flag.String("a", "localhost:8080", "Адрес для запуска сервера (по умолчанию: localhost:8080)")
	returnPrefix := flag.String("b", "http://localhost:8080", "Префикс для возвращаемых сокращённых URL (по умолчанию: http://localhost:8080)")
	logLevel := flag.String("l", "info", "Уровень логирования: debug, info, warn, error")
	storageFilePath := flag.String("f", "./storage.json", "Путь к файлу хранения данных")
	databaseDSN := flag.String("d", "", "DSN")
	auditFile := flag.String("audit-file", "", "Путь к файлу для аудиита")
	auditURL := flag.String("audit-url", "", "URL для аудиита")

	flag.Parse()
	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		runAddr = &envRunAddr
	}
	if envReturnPrefix := os.Getenv("BASE_URL"); envReturnPrefix != "" {
		returnPrefix = &envReturnPrefix
	}
	if envStorageFilePath := os.Getenv("FILE_STORAGE_PATH"); envStorageFilePath != "" {
		storageFilePath = &envStorageFilePath
	}
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		databaseDSN = &envDatabaseDSN
	}
	if envAuditFile := os.Getenv("AUDIT_FILE"); envAuditFile != "" {
		auditFile = &envAuditFile
	}
	if envAuditURL := os.Getenv("AUDIT_URL"); envAuditURL != "" {
		auditURL = &envAuditURL
	}

	var level zapcore.Level
	if err := level.UnmarshalText([]byte(*logLevel)); err != nil {
		level = zapcore.InfoLevel
	}
	logger := logger.Initialize(level)
	return &Config{
		RunAddr:         *runAddr,
		ReturnPrefix:    *returnPrefix,
		Logger:          *logger,
		StorageFilePath: *storageFilePath,
		DatabaseDSN:     *databaseDSN,
		AuditURL:        *auditURL,
		AuditFile:       *auditFile,
	}
}

func NewConfig() *Config {
	return ParseFlags()
}
