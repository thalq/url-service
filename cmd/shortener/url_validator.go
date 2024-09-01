package main

import "net/url"

func ifValidURL(testURL string) bool {
	parsedUrl, err := url.ParseRequestURI(testURL)
	if err != nil {
		return false
	}

	if parsedUrl.Scheme == "" || parsedUrl.Host == "" {
		return false
	}

	return true

}
