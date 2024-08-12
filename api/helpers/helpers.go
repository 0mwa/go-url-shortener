package helpers

import (
	"os"
	"strings"
)

func EnforceHTTP(url string) string {
	if url[:4] != "http" {
		return "http://" + url
	}
	return url
}

func DomainError(url string) bool {
	domain := os.Getenv("DOMAIN")
	normalizedURL := normalizeURL(url)
	return normalizedURL != domain
}

func normalizeURL(url string) string {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "www.")
	return strings.Split(url, "/")[0]
}
