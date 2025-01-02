package handlers

import "testing"

func TestIfValidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "valid url",
			url:  "https://google.com",
			want: true,
		},
		{
			name: "invalid url",
			url:  "google.com",
			want: false,
		},
		{
			name: "invalid url",
			url:  "http:/google.com",
			want: false,
		},
		{
			name: "invalid url",
			url:  "https://google.com:80",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ifValidURL(tt.url); got != tt.want {
				t.Errorf("ifValidURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
