# AI instruction 94

## Cli

### Invoices

`invoice` needs also to interprete the `now()`, `now()-1d`, etc.
Make an utils function global.

Also I want a `-o file.pdf` or `--output file.pdf` for `preview` or `generate` (optional, by default it take the `Content-Disposition` filename if precises otherwise `invoice.pdf`).

### Export

```shell
cwclock export --client <client uuid> --begin 2024-01-15T09:00:00 --end 2024-01-15T12:00:00
cwclock export --client <client uuid> --end 2024-01-15T09:00:00 --to 2024-01-15T12:00:00
```

Manage the optional filter and mandatory args properly (i.e: `--client-id` should be optional).

Same, it needs also to interprete the `now()`, `now()-1d`, etc.

And same, I want a `-o file.pdf` or `--output file.pdf` for `preview` or `generate` (optional, by default it take the `Content-Disposition` filename if precises otherwise `(summary|detailed).pdf`).
