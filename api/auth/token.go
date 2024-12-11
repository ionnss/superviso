package auth

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

func GenerateVerificationToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func GenerateVerificationTokenExpiry() time.Time {
	return time.Now().Add(24 * time.Hour)
}
