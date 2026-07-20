# AI instruction 59

## Invoice's email body

If the client has a purchased order filled, display it in the email like this:

```go
details := fmt.Sprintf("period %s, purchased order %s", period, purchaseOrder)
```

In the parenthesis:

```go
body = template.HTML(fmt.Sprintf(
    `<p>Please find attached invoice <strong>%s</strong> from <strong>%s</strong> (%s)%s</p>`,
    template.HTMLEscapeString(invoiceNumber), template.HTMLEscapeString(orgName), template.HTMLEscapeString(details), suffix,
))
```

Same in french (_bon de commande_ instead of _purchased order_).
