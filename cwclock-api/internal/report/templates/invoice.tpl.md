# Invoice N°{{.InvoiceNumber}}

{{.OrgCity}}, the {{.InvoiceDate}}

Purchase Order: {{.ClientPurchaseOrder}}

| Issuer     | Details           |
| ---------- | ----------------- |
| Name       | {{.OrgName}}      |
| Adress     | {{.OrgAddress}}   |
| SIREN      | {{.OrgSIREN}}     |
| SIRET      | {{.OrgSIRET}}     |
| VAT/TVA IC | {{.OrgVATNumber}} |
| Code NAF   | {{.OrgNAF}}       |
| Date       | {{.InvoiceDate}}  |
| Contact    | {{.OwnerContact}} |

| Customer   | Details            |
| ---------- | ------------------ |
| Name       | {{.ClientName}}    |
| Adress     | {{.ClientAddress}} |
| Contact    | {{.ClientEmail}}   |
| VAT/TVA IC | {{.ClientVATLine}} |

## Object

{{.LineItemsMarkdown}}

| Totals                   | Amount  |
| -------------------------| ------- |
| Total without taxes (HT) | {{.TotalHT}} {{.Currency}} |
| VAT/TVA  | {{.TotalVATLine}} |
| Total with taxes (TTC) | {{.TotalTTC}} {{.Currency}} |
