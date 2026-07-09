# AI instruction 31

## Safe date picker

In invoice or report generation, always check that end date is greater or equals than begin date and display an error toast with no result if it's not the case.

## Export limit

Add a limitation of 5000 entries per exports. If the selected period is too large, the api must return a 400 error with `i18n_code: "export_limit_exceeded"` and the frontend must display a toast with the error message.

Same for invoice.

The value of 5000 must be overrideable by a `CWCLOCK_MAX_REPORT_SIZE` environment variable.

## Invoice

In the invoice generation page, the client select field must be a dropdown autocomplete.
