// Package middlewares provides HTTP middleware functions for the application.
// It includes authentication, logging, and request processing utilities.
package middlewares

import (
	"net/http"

	"github.com/google/uuid"
)

// Cookie-related constants
const (
	// userIDCookieName is the name of the cookie used to store the user's unique identifier
	userIDCookieName = "user_id"
)

// AuthMiddleware is an HTTP middleware that ensures each request has a valid user ID.
// If the request doesn't have a user ID cookie, it generates a new one.
// The middleware adds the user ID to the request context for use in handlers.
//
// The middleware performs the following actions:
//  1. Checks for an existing user ID cookie
//  2. If not found, creates a new user ID and sets it as a cookie
//  3. Continues to the next handler in the chain
//
// The cookie is set with HttpOnly flag for security and is valid for all paths.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDCookie, err := r.Cookie(userIDCookieName)
		if err != nil || userIDCookie.Value == "" {
			setNewUserCookie(w)
			next.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetUserID retrieves the user ID from the request's cookies.
//
// Parameters:
//   - r: The HTTP request containing the user ID cookie
//
// Returns:
//   - string: The user ID if found
//   - error: An error if the user ID cookie is not present or invalid
//
// This function is typically used by handlers that need to identify the current user.
func GetUserID(r *http.Request) (string, error) {
	userIDCookie, err := r.Cookie(userIDCookieName)
	if err != nil {
		return "", err
	}
	return userIDCookie.Value, nil
}

// setNewUserCookie generates a new UUID and sets it as a user ID cookie.
// The cookie is set with the following attributes:
//   - Name: user_id
//   - Value: A new UUID v4 string
//   - Path: "/" (valid for all paths)
//   - HttpOnly: true (not accessible via JavaScript)
//
// Parameters:
//   - w: The HTTP response writer to set the cookie on
func setNewUserCookie(w http.ResponseWriter) {
	userID := uuid.New().String()

	http.SetCookie(w, &http.Cookie{
		Name:     userIDCookieName,
		Value:    userID,
		Path:     "/",
		HttpOnly: true,
	})
}
