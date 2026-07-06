# AI instruction 11

## Swagger/open API

On / endpoint instead of _Hello Welcome to cwclock Backend_ I'd like an openapi/swagger page built dynamically from the router.

## Reports

### Summary

The calculation of each task duration in summary is wrong: it should be the sum of each task having the same label or name (a task is a line in the report and multiple time records having the same name).

Also when a record is flagged as "all day", it's not from 00:00 to 23:59 but only from 9:00 to the number of hours per day set in the client.

### Detailed

Detailed is broken with this error:

```
DetailedReportView.jsx:40 Uncaught TypeError: Cannot read properties of undefined (reading 'length')
    at Yi (DetailedReportView.jsx:40:11)
```

Fix it.

### PDF exports

We miss the avatar/logo organization in the PDF (if the organization doesn't have one, pick the [cwclock logo](./assets/cwclock-logo.png) by default).

## UX/UI

Fields doesn't have the same height like "shortcuts" and date pickers.
I want those to have the same height.
