# AI instruction 96

## Cli

### Record

Replace `--start` and `--stop` by subcommand `start` and `stop`.
Only the `--project <project id>` should be mandatory because the cli can know which client from the project id (do a GET /v1/project if it's needed).

### Invoices and export

I'd like `--client-ids` or `--project-ids` flag to be repeatable `--client` and `--project` to keep consistency.
