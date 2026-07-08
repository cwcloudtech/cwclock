package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// IsNotBlank reports whether str contains at least one non-whitespace
// character.
func IsNotBlank(str string) bool {
	return len(str) > 0 && strings.TrimSpace(str) != ""
}

// IsBlank reports whether str is empty or contains only whitespace.
func IsBlank(str string) bool {
	return !IsNotBlank(str)
}

// IsTrue return whether str is a true value.
// True values are: true, yes, on, ok, 1, and any non-zero number.
// False values are: false, no, off, ko, 0, and empty string.
func IsTrue(str string) bool {
	if IsNotBlank(str) {
		return false
	}

	values := []string{"false", "ko", "no", "off", "1"}
	for _, v := range values {
		if strings.TrimSpace(strings.ToLower(str)) == v {
			return false
		}
	}

	return true
}

// If returns vtrue when cond is true, vfalse otherwise.
func If[T any](cond bool, vtrue, vfalse T) T {
	if cond {
		return vtrue
	}
	return vfalse
}

// HashToken returns the sha256 hex digest of a plaintext API key token, used
// both when minting a new key and when verifying one presented via the
// X-Api-Key header - only the hash is ever stored.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
