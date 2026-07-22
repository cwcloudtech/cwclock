# AI instruction 73

## A4 invoices

I'd like the invoices PDF to be in A4 format.

## Export jobs

I want `export_jobs` (new table) with:
* a cron expr and a helper with shortcuts
* a target
* a checkbox with all type of report (we can send multiple types):
  * summary PDF
  * summary CSV
  * detailed PDF
  * detailed CSV
* a time period understanding `now()` and `now()-1d` or `now()-1h` with helper
* other filters with multiselect/autocomplete dropdown : clients and projects
* a swicth boolean to include financial data or not in the reports

Once the job is created it's scheduled, it will send the generated report(s) to the list of emails.

Export job can be setup and manage (create/update/delete) in an organization by owner or admin.

I want a new screen with icon in the sidebar.

Use a go cron scheduler lib.

### Targets

Target can be:
* a list of emails in `to` and a list of emails in `cc` separated with `,` or `;` as it's already done in the invoice's email field (reuse the same utils function)
* a S3/git or google drive connexion (store the same way it's stored in organization)

We can have multiple targets (as in organization : in the data's payload).
