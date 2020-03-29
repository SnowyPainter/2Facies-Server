package utility

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
	alpha   = []byte{'@'}
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

func SplitHeaderBody(b []byte) (header [][]byte, body []byte) {

	b = bytes.TrimSpace(bytes.Replace(b, newline, space, -1))
	div := bytes.Split(b, alpha)
	header = bytes.Split(div[0], space)
	body = div[1]

	return
}
