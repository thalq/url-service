package operations

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	logger "github.com/thalq/url-service/internal/middleware"
	"github.com/thalq/url-service/internal/structures"
)

const SecretKey = "supersecretkey"

func GetUserID(r *http.Request) (string, error) {
	tokenString, err := r.Cookie("token")
	if err != nil {
		return "", err
	}
	claims := &structures.Claims{}
	token, err := jwt.ParseWithClaims(tokenString.Value, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(SecretKey), nil
		})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		logger.Sugar.Error("Token is not valid")
		return "", err
	}

	logger.Sugar.Infof("Token is valid")
	logger.Sugar.Infof("User ID: %s", claims.UserID)
	return claims.UserID, nil
}
