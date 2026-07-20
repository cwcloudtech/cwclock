# AI instruction 65

## Reports

The calculations of number of billable hours and days is still wrong in the PDF reports and webview. Let-me simplify the rules:

If a time record is flagued as _All day_ I want it to start from `09:00` to `09:00+{number of hours per day}`, no more segments. This way it won't break the UX with edit form in the detailed webview and more simple to compute.

The total number of hours and billable hours should directly add the number of hours per day for each all day time records.

In the report, if multiple clients are selected, for each time records flagued as _All day_ add in the sum the hoursPerDay coming from the associated client, otherwise 7 if it's not set.

And keep the total of days as total of hours divided by hours per day (7 if not set).
