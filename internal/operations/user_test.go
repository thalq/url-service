package operations

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	logger "github.com/thalq/url-service/internal/middleware"
	"github.com/thalq/url-service/internal/models"
)

func TestGetUserID(t *testing.T) {
	logger.InitLogger()

	tests := []struct {
		name          string
		token         string
		expectedID    string
		expectedError bool
	}{
		{
			name:          "Valid Token",
			token:         generateToken(t, "test-user-id"),
			expectedID:    "test-user-id",
			expectedError: false,
		},
		{
			name:          "Invalid Token",
			token:         "invalid-token",
			expectedID:    "",
			expectedError: true,
		},
		{
			name:          "No Token",
			token:         "",
			expectedID:    "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.token != "" {
				cookie := &http.Cookie{
					Name:  "token",
					Value: tt.token,
				}
				req.AddCookie(cookie)
			}

			userID, err := GetUserID(req)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedID, userID)
		})
	}
}

func generateToken(t *testing.T, userID string) string {
	claims := &models.Claims{
		UserID: userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}
	return tokenString
}
func BenchmarkGetUserID(b *testing.B) {
	logger.InitLogger()
	claims := &models.Claims{
		UserID: "test-user-id",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "test",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		b.Fatalf("Failed to sign token: %v", err)
	}

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
