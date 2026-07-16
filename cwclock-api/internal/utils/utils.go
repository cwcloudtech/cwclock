package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"os"
	"regexp"
	"slices"
	"strconv"
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

// SplitNonBlank splits a comma-separated list, trims each entry, and drops
// blanks - so an unset or empty string yields an empty slice rather than
// [""].
func SplitNonBlank(str string) []string {
	var out []string
	for _, part := range strings.Split(str, ",") {
		part = strings.TrimSpace(part)
		if IsNotBlank(part) {
			out = append(out, part)
		}
	}
	return out
}

// GetEnv returns the environment variable key, or fallback when it is unset
// or blank.
func GetEnv(key, fallback string) string {
	v := os.Getenv(key)
	return If(IsNotBlank(v), v, fallback)
}

// IsTrue return whether str is not a false value.
// False values are: false, no, off, ko, 0, and empty string.
func IsTrue(str string) bool {
	if IsBlank(str) {
		return false
	}

	normalized := strings.TrimSpace(strings.ToLower(str))

	falseValues := []string{"false", "ko", "no", "off", "0"}
	if slices.Contains(falseValues, normalized) {
		return false
	}

	if num, err := strconv.ParseFloat(normalized, 64); err == nil {
		return num > 0
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

// emailPattern is a permissive local@domain.tld shape check, not full RFC
// 5322 validation - just enough to catch obvious typos in an optional field.
var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// IsValidEmail reports whether str looks like a plausible email address.
func IsValidEmail(str string) bool {
	return emailPattern.MatchString(str)
}

// ImageSizeExceeds reports whether a base64 (optionally data-URI prefixed)
// image string decodes to more than maxBytes. A blank image never exceeds,
// so clearing a picture/stamp is always allowed regardless of the limit.
func ImageSizeExceeds(image string, maxBytes int64) bool {
	if IsBlank(image) {
		return false
	}
	payload := image
	if strings.HasPrefix(image, "data:") {
		if comma := strings.IndexByte(image, ','); comma >= 0 {
			payload = image[comma+1:]
		}
	}
	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return false
	}
	return int64(len(decoded)) > maxBytes
}
