package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/thalq/url-service/internal/constants"
	"github.com/thalq/url-service/internal/structures"
)

func SetTokenIntoCookie(w http.ResponseWriter, tokenstring string) {
	cookie := &http.Cookie{
		Name:    "token",
		Value:   tokenstring,
		Expires: time.Now().Add(24 * time.Hour),
		Path:    "/",
	}
	http.SetCookie(w, cookie)
}

func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenString, err := r.Cookie("token")
		if err == nil {
			claims := &structures.Claims{}
			token, err := jwt.ParseWithClaims(tokenString.Value, claims,
				func(t *jwt.Token) (interface{}, error) {
					if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
					}
					return []byte(SecretKey), nil
				})
			if err != nil {
				http.Error(w, "Failed to parse token", http.StatusInternalServerError)
				return
			}

			if !token.Valid {
				http.Error(w, "Token is not valid", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), constants.UserIDKey, claims.UserID)
			r = r.WithContext(ctx)
		} else {
			tokenString, userID, err := BuildJWTString()
			if err != nil {
				http.Error(w, "Failed to build JWT string", http.StatusInternalServerError)
				return
			}
			ctx := context.WithValue(r.Context(), constants.UserIDKey, userID)
			r = r.WithContext(ctx)
			SetTokenIntoCookie(w, tokenString)

		}
		next.ServeHTTP(w, r)
	})
}
