package middlewares

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"

	"github.com/google/uuid"
)

const (
	userIDCookieName    = "user_id"
	signatureCookieName = "signature"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDCookie, err := r.Cookie(userIDCookieName)
		fmt.Println(userIDCookie)
		if err != nil || userIDCookie.Value == "" {
			setNewUserCookie(w)
			next.ServeHTTP(w, r)
			return
		}

		signatureCookie, err := r.Cookie(signatureCookieName)
		if err != nil || !verifySignature(userIDCookie.Value, signatureCookie.Value) {
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

	signatureCookie, err := r.Cookie(signatureCookieName)
	if err != nil {
		return "", err
	}

	if !verifySignature(userIDCookie.Value, signatureCookie.Value) {
		return "", http.ErrNoCookie
	}

	return userIDCookie.Value, nil
}

func setNewUserCookie(w http.ResponseWriter) {
	userID := uuid.New().String()
	signature := generateSignature(userID)

	http.SetCookie(w, &http.Cookie{
		Name:     userIDCookieName,
		Value:    userID,
		Path:     "/",
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     signatureCookieName,
		Value:    signature,
		Path:     "/",
		HttpOnly: true,
	})
}

func generateSignature(userID string) string {
	h := hmac.New(sha256.New, []byte(getSecretKey()))
	h.Write([]byte(userID))
	return hex.EncodeToString(h.Sum(nil))
}

func verifySignature(userID, signature string) bool {
	expectedSignature := generateSignature(userID)
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

func getSecretKey() string {
	key := os.Getenv("SECRET_KEY")
	if key == "" {
		return "default-secret-key-change-me-in-production"
	}
	return key
}
