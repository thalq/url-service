package middleware

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/thalq/url-service/internal/structures"
)

const TokenExp = time.Hour * 3
const SecretKey = "supersecretkey"

func BuildJWTString() (string, string, error) {
	userID := uuid.New().String()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, structures.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", "", err
	}

	return tokenString, userID, nil
}
