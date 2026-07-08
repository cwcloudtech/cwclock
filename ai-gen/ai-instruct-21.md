# AI instruction 21

## Time record webview

Paginate the last time records (with the API also) when scrolling with a spinner.

Also print each days as separation with a total of `hh:mm:ss` like the first line.

The `{client} - {project}` in the table must be underlined in the hex color set (add a square circled background as it's done for the user role with the hex color as background color).

## API keys

A user should be able to set an API key to authenticate himself in scripts.

The header will be `X-Api-Key` and will be prior to the JWT token if both are present.

An API key can have an expiration date and a description (add a new table `api_key`).

## Export API

In some export project I have this kind of scripts:

```shell
#!/usr/bin/env bash

#In order to run this script you must use this format
# ./automate-result.sh  dateRangeStart dateRangeEnd emailReceiver
# Example: ./automate-result.sh "2021-06-26" "2021-06-30" "support@nexere.com"

#declaration of variables
dateRangeStart="${1}" #Beginning Date of the Report
dateRangeEnd="${2}" #End Date of the Report
mailReceiver="${3}" #List of emails of the client separated by ",", could be one.
internalComworkReceiver="${4}" #List of emails of the devops team separated by ",", could be one.
mailFrom="${5}" #Mail sender
apiKey="${6}" #clockify API key
sendgrid_api="${7}" #sendgrip API key

organization_id="<uuid>"
client_id="<uuid>"

echo "[INFO][automate-result] All those mails will receive the clockify export: ${mailReceiver}, from: ${dateRangeStart}, to: ${dateRangeEnd}"

echo -ne '\n'
pdf_report="report-${dateRangeStart}-${dateRangeEnd}.pdf"
curl -H "content-type: application/json" -H "X-Api-Key: $apiKey" -X POST -d "{\"exportType\":\"PDF\",\"sortOrder\":\"DESCENDING\",\"dateRangeStart\":\"${dateRangeStart}T06:00:00.000Z\",\"dateRangeEnd\":\"${dateRangeEnd}T20:00:00.000Z\", \"clients\":{\"ids\": [\"${client_id}\"],\"contains\": \"CONTAINS\",\"status\": \"ALL\"}, \"detailedFilter\": {\"page\": 1,\"pageSize\": 100,\"sortColumn\": \"DATE\" }}" https://reports.api.clockify.me/v1/workspaces/${organization_id}/reports/detailed > "${pdf_report}"
echo "[INFO][automate-result] File ready at:"
ls -l "${pdf_report}"

python email-sender.py ${dateRangeStart} ${dateRangeEnd} ${mailReceiver}
```

I want a similar endpoint but with this url instead:

```shell
https://{api}/v1/organizations/${organization_id}/reports/detailed/pdf
https://{api}/v1/organizations/${organization_id}/reports/summary/pdf
```

And:

```shell
https://{api}/v1/organizations/${organization_id}/reports/detailed/csv
https://{api}/v1/organizations/${organization_id}/reports/summary/csv
```

Adapt the existing reports endpoint to match with this interface contract and adapt the frontend.
