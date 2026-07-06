# AI instruction 9

## Currency

Allow to override the list of currency using an environment variables `CWCLOCK_ALLOWED_CURRENCIES` which will contain a json array like this: `["EUR", "USD", "GBP"]`.

The frontend shouldn't maintain a list but should use the list provided by the backend with an endpoint `GET /v1/currencies` (and which is updated if the environment variable is set).

## Avatar in darkmode

For transparent image, i'd like a lighter background color in darkmode for the avatar area (for both organization and user).
