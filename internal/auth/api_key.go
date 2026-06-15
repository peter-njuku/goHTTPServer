package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("No Authorization header found")
	}

	parts := strings.Fields(authHeader)
	if len(parts) < 2 || strings.ToLower(parts[0]) != "apikey" {
		return "", errors.New("Malformed Authorization header")
	}

	return parts[1], nil
}
