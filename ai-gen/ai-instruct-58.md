# AI instruction 56

## Contact form

* If the user is connected, display _Back to time reporting_ instead of _Back to login_
* The CWCloud's api can return the following errors:
  * `429` with `i18n_code: "cf_rate_limiting"` in the body for rate limiting (a user is sending too much emails)
  * `400` with `i18n_code: "message_too_short"` if the message is too short
  * `400` with `i18n_code: "gibberish"` if it looks like spam

I want you to handle those errors and display the appropriate message to the user.

Also, beware `i18n_code` is optional.
