package middlewares

import (
	"net/http"

	"github.com/google/uuid"
)

const (
	userIDCookieName = "user_id"
)

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

func GetUserID(r *http.Request) (string, error) {
	userIDCookie, err := r.Cookie(userIDCookieName)
	if err != nil {
		return "", err
	}
	return userIDCookie.Value, nil
}

func setNewUserCookie(w http.ResponseWriter) {
	userID := uuid.New().String()

	http.SetCookie(w, &http.Cookie{
		Name:     userIDCookieName,
		Value:    userID,
		Path:     "/",
		HttpOnly: true,
	})
}
