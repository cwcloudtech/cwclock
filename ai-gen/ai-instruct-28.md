# AI instruction 28

## Invoice deletion

Add a delete button with confirmation on the table as usual.

## Invoice pdf

I want the value to fit the table cells: the table cells are too small and if the goes out of the table.

If the value is too big, you can use `\n`.

I want also to drop the lines for each client or organization informations if it's blank.

## Invoice calculation

It seems that the total without taxesis wrong: it should be the sum of all price amount of each project or subdivision.

And for each project the price amount should be the daily rate multiply by the `number of hours / client.hourPerDay`.

And for each subdivision if there's subdivision it's `amount of project / number of subdivision`

And finally apply the VAT if it's greater than 0.

## Invoice UX/UI

Keep the same color of report for table headers (bold with `#1cb9f7`).
