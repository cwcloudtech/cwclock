# AI instruction 46

## Move the template and asset directory

Move `report/assets` and `report/templates` directly in `internal` folder.

## Email utils

CWClock will use the email API of CWCloud with two variables:
- `CWCLOUD_API_URL`
- `CWCLOUD_API_KEY`

The endpoint to call is :

```shell
curl -X 'POST' \
  '${CWCLOUD_API_URL}/v1/email' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -H 'X-Auth-Token: ${CWCLOUD_API_KEY}' \
  -d '{
    "from": "cloud@provider.com",
    "to": "recipient@provider.com",
    "bcc": "bcc@provider.com",
    "subject": "Subject",
    "content": "Content",
    "attachment": {
      "mime_type": "application/pdf",
      "file_name": "invoice.pdf",
      "b64": "base64content"
    }
  }'
```

Make an utils to send emails.
Attachment is optional but will be used later.

The utils must act like best effort: if the variables are blank (use `utils.IsBlank`), log that they are missing but do not fail. If CWCloud API is not available, log that it's not available (with the payload) but do not fail.

The content is html, I want a template with the logo of CWClock for all the emails which will be sent. I want a go template file `internal/templates/email.tpl.html`.

The logo is in `internal/assets/cwclock-logo.png` and must be injected as base64 variable in the template with a `{{ .Logo }}` variable.

## Activation mode

I want an environment variable `CWCLOCK_ACTIVATION_MODE` with `admin` as default value.
The allowed values are `admin` or `email`. If it's set to `email` it will send a confirmation link to the user email. If the user clicks on the link, the account will be activated.

Confirmation link must be directly an endpoint of the API like this:

```
/v1/user/confirmation?token={token}
```

The token is a JWT token with the user id and the expiration date.

The disabled user message has to change according to the activation mode.
If the user is disabled the api has to send an `i18n_code` which is different if the activation mode is `admin` or `email` and the frontend has to display a different message according to the `i18n_code`.

## Forgotten password

I want the user to be able to renew it's password if it's forgotten.
The user will send a request to the API with his email. The API will send an email with a link to renew the password (with a frontend form).

## Ban

I want a status `ban` alongside `disabled`, `confirmed` or `superuser`.
It's like `disabled` except the message is explaining that the user has been banned by an administrator (so another `i18n_code` must be sent by the API).

A ban user cannot request a password renewal or confirm his account even if the activation mode is `email` and the confirmation request is not expired.

## Invoice generation

I want a new button which is _Generate with id_ which will open a popin asking for a particular invoice's id.

The webservice `/generate` will take the invoice id as optional field and use it if it's set instead of current computation of this id. If it's not set, continue the same computation rule.

Of course if the requested id already exists with the same id, the webservice will fail with a 409 error and a proper `i18n_code`.

## Invoice sending

On client I want an optional field _Invoice's emails_ which will be the list of emails with `,` or `;` as separator (and trim each element to return a list of emails, I want an utils function for that in `utils.go`).

On the invoice table I want a button for sending the invoice to the client's invoices emails.
If this field is empty, use the client's email instead.

The logo in the email template must be replaced by the organization's avatar if it exists.
