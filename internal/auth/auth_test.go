package auth

import (
	"net/http"
	"testing"
)

func TestGetBearerToken(t *testing.T) {
	headersOk := http.Header{}
	headersOk.Set("Authorization", "Bearer my-secret-jwt-token-string")

	token, err := GetBearerToken(headersOk)
	if err != nil {
		t.Fatalf("Expected no error extracting valid token, got %v", err)
	}
	if token != "my-secret-jwt-token-string" {
		t.Errorf("Expected token to be 'my-secret-jwt-token-string', got %s", token)
	}

	// Test Case 2: Missing header
	headersMissing := http.Header{}
	_, err = GetBearerToken(headersMissing)
	if err == nil {
		t.Errorf("Expected an error when the Authorization header is missing completely")
	}

	// Test Case 3: Malformed layout (missing the "Bearer" prefix keyword)
	headersMalformed := http.Header{}
	headersMalformed.Set("Authorization", "Token my-secret-jwt-token-string")
	_, err = GetBearerToken(headersMalformed)
	if err == nil {
		t.Errorf("Expected an error when the header prefix is not 'Bearer'")
	}
}
