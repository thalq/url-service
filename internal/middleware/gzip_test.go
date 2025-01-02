package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGzipMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	tests := []struct {
		name               string
		requestEncoding    string
		responseEncoding   string
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:               "No Gzip",
			requestEncoding:    "",
			responseEncoding:   "",
			expectedStatusCode: http.StatusOK,
			expectedBody:       "Hello, World!",
		},
		{
			name:               "Gzip Request and Response",
			requestEncoding:    "gzip",
			responseEncoding:   "gzip",
			expectedStatusCode: http.StatusOK,
			expectedBody:       "Hello, World!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.requestEncoding == "gzip" {
				var buf bytes.Buffer
				gz := gzip.NewWriter(&buf)
				_, err := gz.Write([]byte("Hello, World!"))
				assert.NoError(t, err)
				gz.Close()
				req.Body = io.NopCloser(&buf)
				req.Header.Set("Content-Encoding", "gzip")
			}

			if tt.responseEncoding == "gzip" {
				req.Header.Set("Accept-Encoding", "gzip")
			}

			rr := httptest.NewRecorder()
			GzipMiddleware(handler).ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatusCode, rr.Code)

			if tt.responseEncoding == "gzip" {
				assert.Equal(t, "gzip", rr.Header().Get("Content-Encoding"))
				gr, err := gzip.NewReader(rr.Body)
				assert.NoError(t, err)
				defer gr.Close()
				body, err := io.ReadAll(gr)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, string(body))
			} else {
				assert.Equal(t, tt.expectedBody, rr.Body.String())
			}
		})
	}
}
