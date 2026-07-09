# AI instruction 29

## VAT on client

When I tried to update the VAT % to 0 for some client it still shows 20% after edit. 20% is the default value but 0 can be set for some clients (apply the default value only if the client has no VAT or if VAT is lower than 0).

## UX/UI

In the invoice table display the status with the same style of user status with:
* red for `unpaid`
* green for `paid`
* grey for `canceled`
* orange for `refunded`

Add also a filter to the status with a multiselect drop down in the list of filters.

## PDF

If the purchase order is blank do not print the line, add also an empty line after the purchase order if it's not blank.

Reduce the size of the issuer and client `details` column by 50%, no need for this to be so big.
