# AI instruction 45

## OIDC

I observed login with frontend the following redirection:

```
https://github.com/login/oauth/authorize?client_id=Ov23ligGjNkAuUXwqr4j&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fv1%2Foidc%2Fgithub%2Fcallback&response_type=code&scope=read%3Auser+user%3Aemail&state=XXXXXX
```

It should be redirected to the frontend not the API.
Keep this as default redirectURI but if it's the frontend who's making the API call to get this URI redirect to the frontend instead using the environment variable `CWCLOCK_UI_URL` and the path `/oidc/callback` (and check if the frontend router match).

The frontend can pass a query param `?origin=frontend` to the API to indicate it's the frontend which requires the redirectURI.
