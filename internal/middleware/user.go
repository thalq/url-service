package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/thalq/url-service/internal/structures"
)

const TokenExp = time.Hour * 3
const SecretKey = "supersecretkey"

func generateUserID() (string, error) {
	timestamp := time.Now().UnixNano()

	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	userID := fmt.Sprintf("%x-%s", timestamp, hex.EncodeToString(b))
	return userID[:5], nil
}

func BuildJWTString() (string, error) {
	userID, err := generateUserID()
	if err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, structures.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
