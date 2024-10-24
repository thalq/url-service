package cookies

import (
	"fmt"
	"net/http"
	"time"

	"github.com/thalq/url-service/internal/token"
)

func SetTokenIntoCookie(w http.ResponseWriter, tokenstring string) {
	cookie := &http.Cookie{
		Name:    "token",
		Value:   tokenstring,
		Expires: time.Now().Add(24 * time.Hour),
	}
	http.SetCookie(w, cookie)
}
func GetOrSetCookie(w http.ResponseWriter, r *http.Request) string {
	tokenString, err := r.Cookie("token")
	if tokenString == nil || err != nil {
		fmt.Println("Token is nil")
		tokenString, err := token.BuildJWTString()
		if err != nil {
			http.Error(w, "Failed to build JWT string", http.StatusInternalServerError)
		}
		SetTokenIntoCookie(w, tokenString)
		fmt.Println("Setted token into cookie")
		return tokenString
	}
	return tokenString.Value
}
