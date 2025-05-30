package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashSha256 computes the SHA256 hash of a given string and returns its hex representation.
func HashSha256(data string) string {
	hasher := sha256.New()
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
}
