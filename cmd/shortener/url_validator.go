package main

import (
	"fmt"
	"net/url"
)

func ifValidURL(testURL string) bool {
	parsedURL, err := url.ParseRequestURI(testURL)
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!", parsedURL)
	if err != nil {
		return false
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false
	}

	return true

}
