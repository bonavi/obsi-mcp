package repository

import (
	"crypto/sha256"
	"encoding/hex"
)

// hashContent считает хэш содержимого заметки
func hashContent(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}
