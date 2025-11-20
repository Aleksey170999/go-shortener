package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestParseFlags(t *testing.T) {
	// Save original command line arguments and environment variables
	oldArgs := os.Args
	oldEnv := make(map[string]string)

	// Save and clear environment variables
	for _, env := range []string{
		"SERVER_ADDRESS",
		"BASE_URL",
		"FILE_STORAGE_PATH",
		"DATABASE_DSN",
		"AUDIT_FILE",
		"AUDIT_URL",
	} {
		if val, ok := os.LookupEnv(env); ok {
			oldEnv[env] = val
			os.Unsetenv(env)
		}
	}

	// Restore original state when test is done
	defer func() {
		os.Args = oldArgs
		for k, v := range oldEnv {
			os.Setenv(k, v)
		}
	}()

	// Test cases
	testCases := []struct {
		name          string
		args          []string
		envVars       map[string]string
		expected      *Config
		expectedLevel zapcore.Level
	}{
		{
			name:    "default values",
			args:    []string{"cmd"},
			envVars: map[string]string{},
			expected: &Config{
				RunAddr:         "localhost:8080",
				ReturnPrefix:    "http://localhost:8080",
				StorageFilePath: "./storage.json",
				DatabaseDSN:     "", // Default is empty string
				AuditFile:       "",
				AuditURL:        "",
			},
			expectedLevel: zapcore.InfoLevel,
		},
		{
			name: "command line flags",
			args: []string{
				"cmd",
				"-a=:9090",
				"-b=https://example.com",
				"-l=debug",
				"-f=/tmp/storage.json",
				"-d=host=localhost port=5432 user=user password=pass dbname=db sslmode=disable",
				"-audit-file=/tmp/audit.log",
				"-audit-url=http://audit.example.com",
			},
			envVars: map[string]string{},
			expected: &Config{
				RunAddr:         ":9090",
				ReturnPrefix:    "https://example.com",
				StorageFilePath: "/tmp/storage.json",
				DatabaseDSN:     "host=localhost port=5432 user=user password=pass dbname=db sslmode=disable",
				AuditFile:       "/tmp/audit.log",
				AuditURL:        "http://audit.example.com",
			},
			expectedLevel: zapcore.DebugLevel,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tc.envVars {
				t.Setenv(k, v)
			}

			// Reset command line flags
			flag.CommandLine = flag.NewFlagSet(tc.args[0], flag.ExitOnError)

			// Set command line arguments
			os.Args = tc.args

			// Parse flags
			config := ParseFlags()

			// Verify the config values
			assert.Equal(t, tc.expected.RunAddr, config.RunAddr)
			assert.Equal(t, tc.expected.ReturnPrefix, config.ReturnPrefix)
			assert.Equal(t, tc.expected.StorageFilePath, config.StorageFilePath)
			assert.Equal(t, tc.expected.DatabaseDSN, config.DatabaseDSN)
			assert.Equal(t, tc.expected.AuditFile, config.AuditFile)
			assert.Equal(t, tc.expected.AuditURL, config.AuditURL)

			// Verify the logger level
			assert.Equal(t, tc.expectedLevel, config.Logger.Level())
		})
	}
}

func TestNewConfig(t *testing.T) {
	// Save and restore command line args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Reset command line flags
	flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)
	os.Args = []string{"test"}

	// Test that NewConfig doesn't panic and returns a valid config
	config := NewConfig()
	assert.NotNil(t, config)
	assert.NotNil(t, config.Logger)
}
