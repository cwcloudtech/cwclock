# AI instruction 83

## Mail limit

Add an environment variable `CWCLOCK_LIMIT_MAIL` (set to 100 if it's not defined) which will be used for counting the emails sent by organizations that are not owned by a `superuser`.

The concerned emails are only the following:
* invoice emails
* export job emails

If this limit is passed, the mail won't be sent until counter is reset.

Counter will be set in a table like this:

```
mail_counter(orga_id, count, updated_at)
```

If updated_at is older than the current month, count is set to `0`.
