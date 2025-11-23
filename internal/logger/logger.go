// Package logger provides a structured logging implementation using zap logger.
// It offers a pre-configured logger with JSON formatting and configurable log levels.
// The logger is designed to be used throughout the application for consistent logging.
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Initialize creates and configures a new zap.Logger instance with the specified log level.
// The logger is configured with JSON formatting, ISO8601 timestamps, and outputs to stdout/stderr.
//
// Parameters:
//   - level: The minimum log level to output (e.g., zapcore.InfoLevel, zapcore.DebugLevel)
//
// Returns:
//   - *zap.Logger: A configured logger instance
//
// The logger includes the following fields by default in each log entry:
//   - ts: ISO8601 formatted timestamp
//   - level: Log level (debug, info, warn, error, etc.)
//   - caller: Source file and line number of the log call
//   - msg: The actual log message
//
// Note: The function will panic if the logger cannot be created, as logging is considered
// a critical part of the application's operation.
func Initialize(level zapcore.Level) *zap.Logger {
	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Development: false,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	return logger
}
