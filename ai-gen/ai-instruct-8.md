# AI instruction 8

## I18N remaining parts

The remaining parts should also be translated:
* All messages in the toasts
* All error messages from the api (it should add in the payload a `i18n_code` and in the frontend if it's present, it should be translated using i18n system)

The languages in the connected user dropdown should be a select of existing languages in case there's a third one in the future, also with flags representing the language (uk for english, france for french).

## Organization currency

It must be fixed by a list (dropdown autocomplete) of ISO 4217 currency codes (3 chars) and must be the following only in this order:

* `EUR`
* `USD`
* `GBP`
* `CAD`
* `CHF`
* `TND`
* `DZD`
* `MAD`
* `TRY`
* `EGP`
* `SAR`
* `AED`
* `QAR`
* `CNY`
* `HKD`
* `SGD`
* `JPY`
* `AUD`
* `NZD`

By default it's `EUR` for every organization.

Update all forms (creation/edit organization from user and superuser perspective).
