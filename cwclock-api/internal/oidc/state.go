package oidc

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// stateTTL bounds how long a login redirect can take before its state is
// rejected - generous enough for a user dawdling on the provider's consent
// screen, tight enough to keep a captured state parameter useless quickly.
const stateTTL = 10 * time.Minute

// ErrInvalidState is returned by VerifyState for a forged, expired or
// mismatched-provider state parameter.
var ErrInvalidState = errors.New("oidc: invalid or expired state")

// SignState produces a self-contained, tamper-proof state parameter binding
// the login redirect to a specific provider and a short expiry - there's no
// server-side session to store it against, so the signature is what makes it
// trustworthy on the way back in VerifyState.
func SignState(secret, provider string) (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	payload := fmt.Sprintf("%s|%d|%s", provider, time.Now().Unix(), base64.RawURLEncoding.EncodeToString(nonce))
	sig := sign(secret, payload)
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

// VerifyState checks a state parameter returned in a callback was issued by
// SignState for the given provider and hasn't expired.
func VerifyState(secret, provider, state string) error {
	parts := strings.SplitN(state, ".", 2)
	if len(parts) != 2 {
		return ErrInvalidState
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return ErrInvalidState
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ErrInvalidState
	}
	if !hmac.Equal(sig, sign(secret, string(payload))) {
		return ErrInvalidState
	}

	fields := strings.SplitN(string(payload), "|", 3)
	if len(fields) != 3 || subtle.ConstantTimeCompare([]byte(fields[0]), []byte(provider)) != 1 {
		return ErrInvalidState
	}
	issuedAt, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return ErrInvalidState
	}
	if time.Since(time.Unix(issuedAt, 0)) > stateTTL {
		return ErrInvalidState
	}
	return nil
}

func sign(secret, payload string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return mac.Sum(nil)
}
