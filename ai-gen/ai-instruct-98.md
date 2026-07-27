# AI instruction 98

## Cli

### Invoice command

Also `-i <invoice uuid>` if the uuid is not correct or not found try as if it's the invoice name (add an api endpoint `GET /invoices?name=`" if it's required.

Same for the client with (`--client` or `--id` everywhere), search by it's name, same for projects (`--project` or `--id`).

Same for `--org`.

All those case insensitive and pick the first match.
