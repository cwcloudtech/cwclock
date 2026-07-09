# AI instruction 26

## Add missing fields in frontend

* `Organization.NAF` labeled `NAF Code` or `Code NAF` in French (not mandatory, text)
* `Client.email` (not mandatory but check email format like it's supposed to be done for user)

## Invoices

Add a invoice screen in the sidebar for owner or admins of the organization.

It's like the report screen but selecting one client at a time is a required filter in the form and begin date and end date are required with the following shortcut:

* this month
* last month

Then it has to generate pdf invoice with `preview` which will only generate the pdf or `generate` which will generate and save in a `invoices` table (associated with the organization and client).

The invoices table should contain:
* id (uuid)
* data: jsonb
* pdf (blob)
* organization_id (uuid)
* client_id
* created_at
* updated_at
* selected_begin_date
* selected_end_date

Invoice's data should contain:
* id (str)
* status (text: enum {`paid`, `unpaid`, `canceled`, `refunded`}, default `unpaid`)
* totalHT (float)
* totalVAT (float)
* totalTTC (float)

The id is like `{CLIENT_NAME_CAPITALIZED}{YYYYMMDD}{incremental_number}` (if this id already exists in database just add `+1`).

The template should be like this:

```markdown
![logo](../images/logo-{{organization.id}}.png)

<div align="right">
{{ .Organization.city }}, le {{ .CurrentDate.Format "02/01/2006" }}
</div>

# Facture N°{{ .Invoice.id }}

Référence bon de commande (purchase order): {{ .Client.PurchaseOrder }}

| Emetteur | {{ .Organization.name }}                 |
| ---------| --------------------------------------- |
| Adresse  | {{ .Organization.address }}, {{ .Organization.postalCode }} {{ .Organization.city }}, {{ .Organization.country }} |
| N°SIREN  | {{ .Organization.SIREN }}                |
| N°SIRET  | {{ .Organization.SIRET }}                |
| TVA IC   | {{ Organization.VATNumber }}            |
| Code NAF | {{ Organization.NAF }}                  |
| Date     | {{ .Invoice.date.Format "02/01/2006" }} |
| Contact  | {{ .Owner.surname }} {{ .Owner.name }}: {{ .Owner.email }} |

| Client   | {{ .Client.name }}     |
| -------- | ---------------------- |
| Adresse  | {{ .Client.address }}, {{ .Client.postalCode }} {{ .Client.city }}, {{ .Client.country }}  |
| Contact  | {{ .Client.email }}       |
| TVA IC   | {{ .Client.varNumber }} |

## Objet

V_INVOICE_ARRAY

| Totaux    |                         |
| --------- | ----------------------- |
| Total HT  | {{ .Invoice.totalHT }}  |
| TVA       | {{ .Invoice.totalVAT }} ({{ .Client.VATRate }}) |
| Total TTC | {{ .Invoice.totalTTC }} |

![stamp](temporary_assets/stamp-{{ .Organization.id }}.png)
```

Notes:
* `HT` mean `without taxes` and `TTC` mean `with taxes` in French, keep those in english as well.
* `TVA` is `VAT` in French

Rules:

* If {{ .Client.VATRate }} is lower or equal than 0 write _no tva_ (or _pas de tva_) in the cell and if a `.Client.VATDischargeMotive` is not blank, add it like this: `no tva - {{ .Client.VATDischargeMotive }}`
* `V_INVOICE_ARRAY` has to be replaced an array with a line per project like this:

```markdown
| Description | Quantité | Montant HT |
| ----------- | -------  | ---------- |
| {{ .project.name }} | {{ .project.quantity }} | {{ .project.totalHT }} | {{ .Service.totalHT }} |
```

The quantity is the total of hours of the project divided by `.Client.HoursPerDay` (default `7`).

If the project contain subdivision, print a line per subdivision instead of project line and divide for each line the quantity per the number of subdivision.

The webview display also the existing invoices between the two selected dates and allow to download the pdf or update the status with a dropdown if we click on "edit" button.
