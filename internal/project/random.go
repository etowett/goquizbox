package project

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// RandomHexString generates a random string of the provided length.
func RandomHexString(length int) (string, error) {
	b, err := RandomBytes((length + 1) / 2)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b)[:length], nil
}

// RandomBase64String encodes a random base64 string of a given length.
func RandomBase64String(length int) (string, error) {
	b, err := RandomBytes(length)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:length], nil
}

// RandomBytes returns a byte slice of random values of the given length.
func RandomBytes(length int) ([]byte, error) {
	buf := make([]byte, length)
	n, err := rand.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random: %w", err)
	}
	if n < length {
		return nil, fmt.Errorf("insufficient bytes read: %v, expected %v", n, length)
	}
	return buf, nil
}
