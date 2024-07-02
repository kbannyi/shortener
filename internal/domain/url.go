package domain

import (
	"crypto/md5"
	"encoding/hex"
)

type URL struct {
	ID    string
	Value string
}

func NewURL(value string) *URL {

	return &URL{
		ID:    GetMD5Hash(value)[:8],
		Value: value,
	}
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
