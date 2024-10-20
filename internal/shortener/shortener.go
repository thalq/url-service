package shortener

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

func GenerateShortString(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	hashBytes := h.Sum(nil)

	hexString := hex.EncodeToString(hashBytes)

	encodedString := base64.StdEncoding.EncodeToString([]byte(hexString))

	if len(encodedString) > 8 {
		return encodedString[:8]
	}
	return encodedString
}
