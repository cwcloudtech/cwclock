# Invoice N°{{.InvoiceNumber}}

{{.OrgCity}}, the {{.InvoiceDate}}

Purchase Order: {{.ClientPurchaseOrder}}

| Issuer     | Details           |
| ---------- | ----------------- |
| Name       | {{.OrgName}}      |
| Adress     | {{.OrgAddress}}   |
| SIREN      | {{.OrgSIREN}}     |
| SIRET      | {{.OrgSIRET}}     |
| TVA/VAT IC | {{.OrgVATNumber}} |
| Code NAF   | {{.OrgNAF}}       |
| Date       | {{.InvoiceDate}}  |
| Contact    | {{.OwnerContact}} |

| Customer | Details            |
| -------- | ------------------ |
| Name     | {{.ClientName}}    |
| Adress   | {{.ClientAddress}} |
| Contact  | {{.ClientEmail}}   |
| TVA IC   | {{.ClientVATLine}} |

## Object

{{.LineItemsMarkdown}}

| Totals                      | Amount  |
| --------------------------- | ------- |
| Total without taxes (HT)    | {{.TotalHT}} {{.Currency}} |
| TVA/VAT | {{.TotalVATLine}} |
| Total with taxes (TTC)      | {{.TotalTTC}} {{.Currency}} |
