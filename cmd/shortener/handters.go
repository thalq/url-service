package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func PostHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Не удалось прочитать тело запроса", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	bodyLink := string(body)

	newLink := generateShortString(bodyLink)

	URLStorage.Lock()
	URLStorage.m[newLink] = bodyLink
	URLStorage.Unlock()

	fmt.Println("POST: Saved URL:", bodyLink, "with key:", newLink)
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fullURL := r.URL.Scheme + "://" + r.Host + r.RequestURI
	if r.URL.Scheme == "" {
		if r.TLS != nil {
			fullURL = "https://" + r.Host + r.RequestURI
		} else {
			fullURL = "http://" + r.Host + r.RequestURI
		}
	}
	if _, err := w.Write([]byte(fullURL + newLink)); err != nil {
		http.Error(w, "Не удалось записать ответ", http.StatusInternalServerError)
	}
}

func GetHandler(w http.ResponseWriter, r *http.Request) {

	url := strings.TrimPrefix(r.URL.Path, "/")
	URLStorage.RLock()
	originalURL, ok := URLStorage.m[url]
	URLStorage.RUnlock()

	err := ifValidURL(originalURL)
	if !err {
		http.Error(w, "Невалидный URL", http.StatusBadRequest)
		return
	}
	fmt.Println("GET: Requested key:", url)
	if ok {
		fmt.Println("GET: Found URL:", originalURL)
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
		fmt.Println("Temporary Redirect sent for URL:", originalURL)
		return
	} else {
		fmt.Println("GET: Key not found:", url)
		http.Error(w, "URL не найден", http.StatusNotFound)
		return
	}
}