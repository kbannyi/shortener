package domain

import (
	"crypto/md5"
	"encoding/hex"
)

type URL struct {
	ID       string `json:"uuid" db:"id"`
	Short    string `json:"short_url" db:"short_url"`
	Original string `json:"original_url" db:"original_url"`
}

func NewURL(value string) *URL {
	hash := GetMD5Hash(value)[:8]
	return &URL{
		ID:       hash,
		Short:    hash,
		Original: value,
	}
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
