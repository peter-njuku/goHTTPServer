package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func MakeRefreshToken() string {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate token: %s", err))
	}
	return hex.EncodeToString(key)
}
