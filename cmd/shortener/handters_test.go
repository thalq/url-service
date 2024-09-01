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
			body:         "https://google.com",
			expectedCode: http.StatusTemporaryRedirect,
		},
		{
			name:         "GetHandlerWrongURL",
			method:       http.MethodGet,
			body:         "google.com",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "GetHandlerMethodNotAllowed",
			method:       http.MethodHead,
			body:         "https://google.com",
			expectedCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.method == http.MethodPost {
				req := httptest.NewRequest(tc.method, "/", strings.NewReader(tc.body))
				w := httptest.NewRecorder()
				PostHandler(w, req)
				assert.Equal(t, tc.expectedCode, w.Code)
				assert.Contains(t, w.Body.String(), "http://")
				assert.Equal(t, tc.contentType, w.Header().Get(tc.header))
			} else if tc.method == http.MethodGet {
				shortURL := generateShortString(tc.body)
				URLStorage.Lock()
				URLStorage.m[shortURL] = tc.body
				URLStorage.Unlock()
				req := httptest.NewRequest(tc.method, "/"+generateShortString(tc.body), nil)
				w := httptest.NewRecorder()
				GetHandler(w, req)
				assert.Equal(t, tc.expectedCode, w.Code)
			}
		})
	}
}
