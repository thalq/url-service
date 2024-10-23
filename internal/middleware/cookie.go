package middleware

import (
	"net/http"
	"time"
)

func SetTokenIntoCookie(w http.ResponseWriter, tokenstring string) {
	cookie := &http.Cookie{
		Name:    "token",
		Value:   tokenstring,
		Expires: time.Now().Add(24 * time.Hour),
	}
	http.SetCookie(w, cookie)
}

func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("token")
		if err != nil {
			tokenString, err := BuildJWTString()
			if err != nil {
				http.Error(w, "Failed to build JWT string", http.StatusInternalServerError)
				return
			}
			SetTokenIntoCookie(w, tokenString)
		}
		next.ServeHTTP(w, r)
	})
}
