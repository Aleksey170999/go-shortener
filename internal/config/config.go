package config

import (
	"flag"
	"os"

	"github.com/Aleksey170999/go-shortener/internal/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds the application configuration parameters.
// It supports configuration via command-line flags and environment variables.
// Environment variables take precedence over command-line flags.
type Config struct {
	RunAddr         string     `env:"SERVER_ADDRESS"` // Server address in format "host:port"
	ReturnPrefix    string     `env:"BASE_URL"`       // Base URL for shortened URLs
	Logger          zap.Logger // Logger instance for application logging
	StorageFilePath string     // Path to file-based storage
	DatabaseDSN     string     // Database connection string
	AuditURL        string     // Remote URL for audit logging
	AuditFile       string     // File path for local audit logging
}

// ParseFlags initializes and parses command-line flags and environment variables.
// It returns a Config struct populated with the parsed values.
// The function follows this precedence order for configuration:
// 1. Environment variables (highest precedence)
// 2. Command-line flags
// 3. Default values (lowest precedence)
//
// Supported environment variables:
//   - SERVER_ADDRESS: Server address (e.g., "localhost:8080")
//   - BASE_URL: Base URL for shortened URLs
//   - FILE_STORAGE_PATH: Path to file storage
//   - DATABASE_DSN: Database connection string
//   - AUDIT_FILE: Path to audit log file
//   - AUDIT_URL: Remote audit service URL
//
// Command-line flags (with their default values):
//   - -a: Server address (default: "localhost:8080")
//   - -b: Base URL (default: "http://localhost:8080")
//   - -l: Log level (default: "info")
//   - -f: Storage file path (default: "./storage.json")
//   - -d: Database DSN (default: empty)
//   - -audit-file: Audit file path (default: empty)
//   - -audit-url: Audit service URL (default: empty)
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

// NewConfig creates and returns a new Config instance by parsing flags and environment variables.
// This is a convenience function that simply calls ParseFlags() and returns its result.
// It's provided for better API ergonomics when a new Config instance is needed.
func NewConfig() *Config {
	return ParseFlags()
}
