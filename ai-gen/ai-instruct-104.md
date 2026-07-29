# AI instruction 104

## Half days

On the same model of "_all day_" I want a boolean `half` in the time record data payload.

On the time record screen and export add another checkbox or switch but when _all day_ is _on_ half is disabled and the other way arround.

On the export or invoice calculation, same rule, apply from 9am to the number of hour of day from the client (or 7 by default) divided by two.

Do not modify the shortcut datepicker for all day. If it's half a day, the user will edit and change the option checked.

Add a `--half` option to the `cwclock record` command (same `--half` and `--all-day` cannot be passed at the same time).
