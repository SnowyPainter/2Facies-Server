package utility

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"

)

func RandomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func DeleteFirst(t []byte, e []byte) []byte {
	f := bytes.Index(t, e) + 1
	result := make([]byte, len(t)-f)

	for i := f; i < len(t); i++ {
		result[i-f] = t[i]
	}
	return result
}
