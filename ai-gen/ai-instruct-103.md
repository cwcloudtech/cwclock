# AI instruction 103

## Api keys

For the following endpoints:

```
GET /me/config/file
GET /me/config/qr
```

I want you to replace the `X-Config-Key` header by a body to avoid adding a CORS exception in our reverse proxy

```
POST /me/config/file
{"key": "value"}

POST /me/config/qr
{"key": "value", "orga_id": "uuid"}
```

`orga_id` is still optional.

