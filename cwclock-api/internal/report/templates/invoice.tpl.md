# Facture N°{{.InvoiceNumber}}

{{.OrgCity}}, le {{.InvoiceDate}}

Référence bon de commande (purchase order) : {{.ClientPurchaseOrder}}

| Emetteur | {{.OrgName}} |
| -------- | ------------ |
| Adresse  | {{.OrgAddress}} |
| N°SIREN  | {{.OrgSIREN}} |
| N°SIRET  | {{.OrgSIRET}} |
| TVA IC   | {{.OrgVATNumber}} |
| Code NAF | {{.OrgNAF}} |
| Date     | {{.InvoiceDate}} |
| Contact  | {{.OwnerContact}} |

| Client   | {{.ClientName}} |
| -------- | ---------------- |
| Adresse  | {{.ClientAddress}} |
| Contact  | {{.ClientEmail}} |
| TVA IC   | {{.ClientVATLine}} |

## Objet

{{.LineItemsMarkdown}}
| Totaux                    |                             |
| -------------------------- | --------------------------- |
| Total HT (without taxes)  | {{.TotalHT}} {{.Currency}} |
| TVA (VAT)                 | {{.TotalVATLine}} |
| Total TTC (with taxes)    | {{.TotalTTC}} {{.Currency}} |
