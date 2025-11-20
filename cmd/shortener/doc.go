// Command shortener is the main server application for the URL shortener service.
// It provides HTTP endpoints for shortening URLs, resolving them, and managing user URLs.
//
// Features:
//   - URL shortening via HTTP POST requests
//   - URL redirection via shortened URLs
//   - Support for both in-memory and PostgreSQL storage
//   - Request logging and compression
//   - User authentication
//   - Batch operations
//   - Audit logging to file or remote service
//
// Configuration:
// The service can be configured using environment variables or command-line flags.
// Key configuration options include:
//   - SERVER_ADDRESS: Server address (default: localhost:8080)
//   - BASE_URL: Base URL for shortened links (default: http://localhost:8080)
//   - FILE_STORAGE_PATH: Path to file storage (optional)
//   - DATABASE_DSN: PostgreSQL connection string (optional)
//   - ENABLE_HTTPS: Enable HTTPS (default: false)
//
// Example usage:
//
//	$ go run cmd/shortener/main.go
//	$ SERVER_ADDRESS=:8080 BASE_URL=http://localhost:8080 go run cmd/shortener/main.go
//
// API Endpoints:
//   - POST / - Create a new short URL
//   - GET /{id} - Redirect to the original URL
//   - GET /api/user/urls - Get all URLs for the current user
//   - POST /api/shorten - Create a short URL (JSON API)
//   - POST /api/shorten/batch - Create multiple short URLs in a batch
//   - DELETE /api/user/urls - Delete URLs in batch
//   - GET /ping - Health check endpoint
package main
