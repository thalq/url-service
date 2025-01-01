package operations

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	logger "github.com/thalq/url-service/internal/middleware"
	"github.com/thalq/url-service/internal/models"
)

func BenchmarkGetUserID(b *testing.B) {
	logger.InitLogger()
	claims := &models.Claims{
		UserID: "test-user-id",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "test",
		},
	}

	// Создание JWT токена
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		b.Fatalf("Failed to sign token: %v", err)
	}

	// Создаем тестовый HTTP запрос с токеном в куки
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: tokenString,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetUserID(req)
		if err != nil {
			b.Fatalf("Failed to get user ID: %v", err)
		}
	}
}
