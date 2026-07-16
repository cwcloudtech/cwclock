# AI instruction 44

## OIDC

I want the following environment variable to be used:

```shell
CWCLOCK_OIDC_GOOGLE_CLIENT_ID=""
CWCLOCK_OIDC_GOOGLE_CLIENT_SECRET=""
CWCLOCK_IDC_GITHUB_CLIENT_ID=""
CWCLOCK_OIDC_GITHUB_CLIENT_SECRET=""
CWCLOCK_OIDC_KEYCLOAK_BASE_URL=""
CWCLOCK_OIDC_KEYCLOAK_CLIENT_ID=""
CWCLOCK_OIDC_KEYCLOAK_CLIENT_SECRET=""
CWCLOCK_OIDC_KEYCLOAK_GROUPS=""
```

I want you to implement OIDC connection with those provider (google, github or keycloak) if those variables are set.

Add also an unauthenticated endpoint `GET /v1/oidc` that returns the list of OIDC providers configured.

The endpoint should return a json like this:

```json
{
  "providers": [
    "google",
    "github",
    "keycloak"
  ]
}
```

The frontend will display the login/signup buttons with those providers calling this endpoint.

In the buttons I want also the logo/icon inside.

