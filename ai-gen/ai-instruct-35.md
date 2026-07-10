# AI instruction 35

## Countries

I want the API to provide a GET /v1/countries endpoints with a list of countries.

The response should be a JSON object with the following structure:

```json
{
  "countries": [
    {
      "name": "United States",
      "iso": "US"
    },
    {
      "name": "Canada",
      "iso": "CA"
    }
  ]
}
```

The clients and organizations should use the iso code inside their's data payload instead of free text.

Use in the front a dropdown autocomplete.

Apply also a SQL migration mapping all countries to their iso code, example:

France -> FR
france -> FR
FRANCE -> FR


The migration should be idempotent.

For client and organizations, this field is required.

## Organizations and client's identifications

If the selected country is `FR`, those fields appear:

* SIRET
* SIREN
* NAF

If it's TN:

* MF (meaning "Matricule Fiscale")

For the rest, by default put "Identification number".

VAT Code is for every EU's country.

I want a desision table to know which field to display with an API:

`/v1/fields?country=FR`

```json
{
  "fields": [
    "SIRET",
    "SIREN",
    "NAF",
    "VAT Code"
  ]
}
```

## Currencies

I want you to replace the current `CWCLOCK_ALLOWED_CURRENCIES` by a currencies table.

The default currency should be selected according to the country but let the user decide (i.e: France is EUR but the user can select USD).

## Database model

Here's the database model I'm expecting

```
currencies(is_code varchar(3) primary key, name varchar(255))
countries(iso_code varchar(2) primary key, name varchar(255) not null, currency_iso_code varchar(3) references currencies(is_code));
fields(uuid, is_code varchar(2) references countries(iso_code), name varchar(255));
```

Indexes on iso codes.
This will be easier to maintain.

The frontend is updating the form dynamically when the country is selected.
