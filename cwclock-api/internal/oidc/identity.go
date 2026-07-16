package oidc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// httpClient is package-level so it can share connection pooling/timeouts
// across every provider call; none of these requests carry credentials that
// need per-call isolation.
var httpClient = &http.Client{Timeout: 10 * time.Second}

// Identity is the normalized profile cwclock cares about, regardless of
// which provider it came from.
type Identity struct {
	Email   string
	Name    string
	Surname string
	// Groups is only populated for Keycloak (when a group-membership
	// mapper adds "groups" to the userinfo response); other providers
	// leave it empty.
	Groups []string
}

// ExchangeCode swaps an authorization code for an access token at the
// provider's token endpoint.
func ExchangeCode(ctx context.Context, p Provider, code, redirectURI string) (string, error) {
	form := url.Values{
		"client_id":     {p.ClientID},
		"client_secret": {p.ClientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var payload struct {
		AccessToken      string `json:"access_token"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("oidc: decoding token response: %w", err)
	}
	if payload.Error != "" {
		return "", fmt.Errorf("oidc: token exchange failed: %s (%s)", payload.Error, payload.ErrorDescription)
	}
	if payload.AccessToken == "" {
		return "", errors.New("oidc: token exchange returned no access_token")
	}
	return payload.AccessToken, nil
}

// FetchIdentity retrieves the authenticated user's profile from the
// provider, using the given access token.
func FetchIdentity(ctx context.Context, p Provider, accessToken string) (Identity, error) {
	if p.Name == Github {
		return fetchGithubIdentity(ctx, accessToken)
	}
	return fetchStandardIdentity(ctx, p, accessToken)
}

// fetchStandardIdentity handles Google and Keycloak, both of which expose a
// standard OIDC userinfo endpoint.
func fetchStandardIdentity(ctx context.Context, p Provider, accessToken string) (Identity, error) {
	var claims map[string]any
	if err := getJSON(ctx, p.UserInfoURL, accessToken, "Bearer", &claims); err != nil {
		return Identity{}, err
	}

	email, _ := claims["email"].(string)
	if email == "" {
		return Identity{}, errors.New("oidc: provider did not return an email")
	}

	given, _ := claims["given_name"].(string)
	family, _ := claims["family_name"].(string)
	if given == "" && family == "" {
		full, _ := claims["name"].(string)
		given, family = splitName(full)
	}

	var groups []string
	if raw, ok := claims["groups"].([]any); ok {
		for _, g := range raw {
			if s, ok := g.(string); ok {
				groups = append(groups, s)
			}
		}
	}

	return Identity{Email: email, Name: given, Surname: family, Groups: groups}, nil
}

func fetchGithubIdentity(ctx context.Context, accessToken string) (Identity, error) {
	var profile struct {
		Login string `json:"login"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := getJSON(ctx, "https://api.github.com/user", accessToken, "Bearer", &profile); err != nil {
		return Identity{}, err
	}

	email := profile.Email
	if email == "" {
		var err error
		email, err = fetchGithubPrimaryEmail(ctx, accessToken)
		if err != nil {
			return Identity{}, err
		}
	}
	if email == "" {
		return Identity{}, errors.New("oidc: github account has no accessible verified email")
	}

	name, surname := splitName(profile.Name)
	if name == "" && surname == "" {
		name = profile.Login
	}
	return Identity{Email: email, Name: name, Surname: surname}, nil
}

// fetchGithubPrimaryEmail covers accounts whose email is private: GitHub
// then omits it from /user and it must be read from /user/emails instead,
// picking the primary, verified address (falling back to any verified one).
func fetchGithubPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := getJSON(ctx, "https://api.github.com/user/emails", accessToken, "Bearer", &emails); err != nil {
		return "", err
	}

	var verified string
	for _, e := range emails {
		if !e.Verified {
			continue
		}
		if e.Primary {
			return e.Email, nil
		}
		if verified == "" {
			verified = e.Email
		}
	}
	return verified, nil
}

func getJSON(ctx context.Context, endpoint, token, scheme string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", scheme+" "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("oidc: %s returned %d: %s", endpoint, resp.StatusCode, string(body))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// splitName splits a single display name into a first/last name pair on the
// first whitespace run, since our User model stores them separately but
// GitHub (and a name-only OIDC claim) only provides one field.
func splitName(full string) (name, surname string) {
	full = strings.TrimSpace(full)
	if full == "" {
		return "", ""
	}
	parts := strings.SplitN(full, " ", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.TrimSpace(parts[1])
}
