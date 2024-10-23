package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Claims — структура утверждений, которая включает стандартные утверждения
// и одно пользовательское — UserID
type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

const TOKEN_EXP = time.Hour * 3
const SECRET_KEY = "supersecretkey"

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
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
