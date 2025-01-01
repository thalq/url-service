package shortener

import "testing"

func BenchmarkShortener(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateShortString("https://www.google.com")
	}
}
