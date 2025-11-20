// Command client provides a simple command-line interface for the URL shortener service.
// It allows users to interact with the shortener service by sending HTTP requests.
// The client sends a POST request to the shortener service with a long URL and receives a shortened version.
//
// Usage:
//  1. Run the client
//  2. Enter a long URL when prompted
//  3. The client will display the server's response, which includes the shortened URL
//
// Example:
//
//	$ go run cmd/client/main.go
//	Введите длинный URL
//	https://example.com/very/long/url
//	Статус-код 201
//	http://localhost:8080/abc123
package main
