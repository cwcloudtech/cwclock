// Package oidc implements the minimal parts of the OAuth2/OIDC
// "authorization code" flow cwclock needs to let a user sign in through
// Google, GitHub or a Keycloak realm: building the authorization redirect,
// exchanging the returned code for an access token, and fetching the
// identity (email/name/groups) it belongs to.
//
// There's no server-side session store, so the flow stays stateless: the
// state parameter is an HMAC-signed, self-contained token (see state.go)
// rather than a lookup key, and the user's identity is read from each
// provider's userinfo endpoint with the access token rather than by
// verifying a signed ID token - simpler than JWKS verification, and
// sufficient since the token was already fetched directly from the
// provider over TLS.
package oidc

import (
	"net/url"

	"cwclock-api/internal/config"
	"cwclock-api/internal/utils"
)

// Provider names, used both as the /v1/oidc list entries and as the
// {provider} route segment for /v1/oidc/{provider}/login|callback.
const (
	Google   = "google"
	Github   = "github"
	Keycloak = "keycloak"
)

type Provider struct {
	Name         string
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	Scope        string
}

// BuildProviders returns the providers enabled by cfg, in a fixed order.
// A provider is only included once all the environment variables it needs
// are set, so an operator can enable any subset of google/github/keycloak
// independently.
func BuildProviders(cfg config.Config) []Provider {
	var providers []Provider

	if utils.IsNotBlank(cfg.OIDCGoogleClientID) && utils.IsNotBlank(cfg.OIDCGoogleClientSecret) {
		providers = append(providers, Provider{
			Name:         Google,
			ClientID:     cfg.OIDCGoogleClientID,
			ClientSecret: cfg.OIDCGoogleClientSecret,
			AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:     "https://oauth2.googleapis.com/token",
			UserInfoURL:  "https://openidconnect.googleapis.com/v1/userinfo",
			Scope:        "openid email profile",
		})
	}

	if utils.IsNotBlank(cfg.OIDCGithubClientID) && utils.IsNotBlank(cfg.OIDCGithubClientSecret) {
		providers = append(providers, Provider{
			Name:         Github,
			ClientID:     cfg.OIDCGithubClientID,
			ClientSecret: cfg.OIDCGithubClientSecret,
			AuthURL:      "https://github.com/login/oauth/authorize",
			TokenURL:     "https://github.com/login/oauth/access_token",
			UserInfoURL:  "https://api.github.com/user",
			Scope:        "read:user user:email",
		})
	}

	if utils.IsNotBlank(cfg.OIDCKeycloakBaseURL) && utils.IsNotBlank(cfg.OIDCKeycloakClientID) && utils.IsNotBlank(cfg.OIDCKeycloakClientSecret) {
		providers = append(providers, Provider{
			Name:         Keycloak,
			ClientID:     cfg.OIDCKeycloakClientID,
			ClientSecret: cfg.OIDCKeycloakClientSecret,
			AuthURL:      cfg.OIDCKeycloakBaseURL + "/protocol/openid-connect/auth",
			TokenURL:     cfg.OIDCKeycloakBaseURL + "/protocol/openid-connect/token",
			UserInfoURL:  cfg.OIDCKeycloakBaseURL + "/protocol/openid-connect/userinfo",
			Scope:        "openid email profile",
		})
	}

	return providers
}

// Find returns the enabled provider with the given name, or false if it
// isn't configured.
func Find(providers []Provider, name string) (Provider, bool) {
	for _, p := range providers {
		if p.Name == name {
			return p, true
		}
	}
	return Provider{}, false
}

// Names returns the enabled providers' names, in the same order as
// BuildProviders, for the GET /v1/oidc response.
func Names(providers []Provider) []string {
	names := make([]string, len(providers))
	for i, p := range providers {
		names[i] = p.Name
	}
	return names
}

// AuthorizationURL builds the URL to redirect the browser to in order to
// start the login with this provider.
func (p Provider) AuthorizationURL(redirectURI, state string) string {
	q := make(url.Values)
	q.Set("client_id", p.ClientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", p.Scope)
	q.Set("state", state)
	return p.AuthURL + "?" + q.Encode()
}
