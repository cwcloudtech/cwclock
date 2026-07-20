# AI instruction 54

## Contact form

I want a contact form which will use the following endpoint of CWCloud:

```shell
curl -X 'POST' \
  "${CWCLOUD_API_URL}/v1/contactreq" \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "id": "<contact form uuid>",
    "email": "foo@bar.com",
    "subject": "Your subjet",
    "message": "Your message",
    "name": "Your name",
    "firstname": "Your first name"
  }'
```

The uuid will be kept in a `CWCLOUD_CONTACT_FORM_ID` environment variable on the backend side.
If this variable is not set, a 405 error will be sent with a proper `i18n_code`.

Notes:
* `name` and `firstname` are optionals
* No `X-Api-Key` header for this particular endpoint
