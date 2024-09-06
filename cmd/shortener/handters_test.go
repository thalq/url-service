package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandlers(t *testing.T) {
	testCases := []struct {
		name         string
		method       string
		body         string
		shortURL     string
		insertURL    string
		header       string
		contentType  string
		expectedCode int
	}{
		{
			name:         "PostHandler",
			method:       http.MethodPost,
			body:         "https://google.com",
			header:       "content-type",
			contentType:  "text/plain",
			expectedCode: http.StatusCreated,
		},
		{
			name:         "GetHandler",
			method:       http.MethodGet,
			insertURL:    "https://test1.com",
			expectedCode: http.StatusTemporaryRedirect,
		},
		{
			name:         "Not-valid URL",
			method:       http.MethodPost,
			body:         "notvalidurl",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Non-Existent URL",
			method:       http.MethodGet,
			shortURL:     "nonexist",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "GetHandlerMethodNotAllowed",
			method:       http.MethodHead,
			body:         "https://test.com",
			expectedCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.method == http.MethodPost {
				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.body))
				w := httptest.NewRecorder()
				PostHandler(w, req)
				assert.Equal(t, tc.expectedCode, w.Code)
				if tc.expectedCode == http.StatusCreated {
					assert.Contains(t, w.Body.String(), "http://")
					assert.Equal(t, tc.contentType, w.Header().Get(tc.header))
				}
			} else if tc.method == http.MethodGet {
				if tc.shortURL != "" {
					req := httptest.NewRequest(http.MethodGet, "/"+tc.shortURL, nil)
					w := httptest.NewRecorder()
					GetHandler(w, req)
					assert.Equal(t, tc.expectedCode, w.Code)
				} else {
					shortURL := generateShortString(tc.insertURL)
					URLStorage.Lock()
					URLStorage.m[shortURL] = tc.insertURL
					URLStorage.Unlock()
					tc.shortURL = shortURL
					req := httptest.NewRequest(http.MethodGet, "/"+tc.shortURL, nil)
					w := httptest.NewRecorder()
					GetHandler(w, req)
					assert.Equal(t, tc.expectedCode, w.Code)
					assert.Equal(t, tc.insertURL, w.Header().Get("Location"))
				}
			}
		})
	}
}
