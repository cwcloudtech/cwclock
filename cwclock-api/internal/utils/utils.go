package utils

import "strings"

// IsNotBlank reports whether str contains at least one non-whitespace
// character.
func IsNotBlank(str string) bool {
	return len(str) > 0 && strings.TrimSpace(str) != ""
}

// IsBlank reports whether str is empty or contains only whitespace.
func IsBlank(str string) bool {
	return !IsNotBlank(str)
}

// If returns vtrue when cond is true, vfalse otherwise.
func If[T any](cond bool, vtrue, vfalse T) T {
	if cond {
		return vtrue
	}
	return vfalse
}
