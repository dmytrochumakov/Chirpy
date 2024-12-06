package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func MakeRefreshToken() (string, error) {
	c := 10
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	encodedStr := hex.EncodeToString(b)
	return encodedStr, nil
}
