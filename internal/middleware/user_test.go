package middleware

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/thalq/url-service/internal/models"
)

func TestBuildJWTString(t *testing.T) {
	tokenString, userID, err := BuildJWTString()
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)
	assert.NotEmpty(t, userID)

	token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	assert.NoError(t, err)
	assert.True(t, token.Valid)

	claims, ok := token.Claims.(*models.Claims)
	assert.True(t, ok)
	assert.Equal(t, userID, claims.UserID)
	assert.WithinDuration(t, time.Now().Add(TokenExp), claims.ExpiresAt.Time, time.Second)
}

func BenchmarkBuildJWTString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, err := BuildJWTString()
		if err != nil {
			b.Fatal(errors.New("Error in BuildJWTString"))
		}
	}
}
