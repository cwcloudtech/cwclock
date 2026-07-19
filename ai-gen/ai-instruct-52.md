# AI instruction 52

## CWCloud's email API

Now it's a list of attachments, not a single attachment.

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
    "attachments": [{
      "mime_type": "application/pdf",
      "file_name": "invoice.pdf",
      "b64": "base64content"
    }]
  }'
```

## Send also the reports

I want a flag option on the client to send also the reports with the invoices (summary and detailed) with the same time periode.

Use the multiple attachements API to send the reports with the invoices.
