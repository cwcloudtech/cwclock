# AI instruction 51

## Email invoice

Add the time period in parenthesis in the title (from/to).

If the client is registered in one of those french speaking countries, translate the email in french:
  * France
  * Belgium
  * Tunisia
  * Algeria
  * Marocco

I want those configuration to be in a decision table:

```
client_language(country_iso, lng_iso)
```

If not present in this table, keep english.
