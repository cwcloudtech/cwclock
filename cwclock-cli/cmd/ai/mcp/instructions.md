# cwclock MCP Usage Guide

Use `list_cwclock_commands` or `get_cwclock_command_help` to discover available commands.
Prefer dynamic tools named `cwclock_<command_path>` for direct execution.

## Natural Language To CLI Flags

- `format`, `output format`, `in <format>`, `en <format>` -> `-f <format>` (format can be `json` or `pretty`)
- `filter`, `label filter` -> `--filter <label:value>` (metrics get only)
- `org`, `organization` -> `-o <org-id>` (most commands; NOT invoice preview/generate or export, see below)
- `client` -> `-c <client-id>` (record start/create, invoice, export)
- `project` -> `-p <project-id>` (record start/create; job create/update, invoice, export accept it repeated)
- `text`, `description` -> `-t <text>` (record start/stop/create)
- `record id` -> `-i <record-id>` (record delete)
- `job id` -> `-i <job-id>` (job run/update/delete/target)
- `invoice id` -> `-i <invoice-id>` (invoice send/upload/delete)
- `begin`, `from`, `start date` -> `--begin <date>` (or its alias `--from`; `--from` only exists on `record`)
- `end`, `to`, `end date` -> `--end <date>` (or its alias `--to`)
- `output file`, `save to`, `save as` -> `-o <path>` / `--output <path>` (invoice preview/generate, export only - `-o` means something different here than on every other command, see below)
- `report type`, `kind of report` -> `--type summary|detailed` (export only, default `summary`)
- `file format` -> `--file-format pdf|csv` (export only, default `pdf`)

## Important Rules

- Dates for `--begin`/`--end` (and their `--from`/`--to` aliases on `record`, or just `--to` on
  `invoice`/`export`) accept ISO-8601, or a relative `now()` style expression such as `now()`,
  `now()-1h`, `now()-1d`. Same parser everywhere (`record create`/`start`, `invoice
  preview`/`generate`, `export`).
- `--org` falls back to the configured `org_id` when omitted; if none is available, the command
  reports that it's missing.
- On `record start` and `record create`, only `--project` is required - `--client` is optional and
  is looked up automatically from the project when omitted. `--text` is optional everywhere in the
  `record` command tree and defaults to the project's name when left blank (matching the web app).
  `record stop` takes no required flags at all - it automatically uses whatever was captured by the
  most recent `record start`.
- `--client`/`--project` are **repeatable** flags (pass `--client id1 --client id2` for more than
  one), not comma-separated, on `job create`/`job update`, `invoice preview`/`generate` (project
  only - an invoice is always for one client), and `export`. Omitting them means "every
  client/project".
- `-o` means `--org` on almost every command, **except** `invoice preview`, `invoice generate` and
  `export`, where `-o` means `--output` (the file path to save the downloaded PDF/CSV to) - on
  those three, `--org` has no short flag.
- `invoice generate` and `export` actually download a file; when the user doesn't give
  `--output`/`-o`, the command still saves it (using the server's suggested filename, or a sane
  default like `invoice.pdf`/`summary.pdf`) and prints the path it wrote to.
- `--org`, `--client`, `--project` (everywhere above) and invoice's `-i`/`--id` all accept either
  the real id **or** the resource's name (invoice's `-i` accepts its number instead of a name) -
  case-insensitively, first match wins. So if the user gives a name instead of a uuid (e.g. "for
  client Acme" or "org Idriss's Org"), just pass it straight through as the flag value - no need to
  look up the real id first. Same idea for `cwclock admin user`'s `-i`/`--id`: it accepts an email
  as a fallback when the value isn't a valid/found user id.
- Whenever both `--client` and `--project` are given together (any command), the command fails
  fast if the resolved project doesn't actually belong to the resolved client - don't try to work
  around that error by guessing a different id, surface it to the user instead, since it means the
  project/client pairing itself was wrong.

## Examples

- `list all metrics` -> `cwclock metrics ls`
- `get metric cpu_all filtered by instance node-1` -> `cwclock metrics get cpu_all --filter instance:node-1`
- `start a timer for working on the CLI on project abc` -> `cwclock record start -t "working on the CLI" -p abc`
- `start a timer on project abc` (no description) -> `cwclock record start -p abc`
- `stop the timer` -> `cwclock record stop`
- `log time from an hour ago until now for a meeting on project abc` -> `cwclock record create --begin now()-1h --end now() -t "meeting" -p abc`
- `list the last 10 records` -> `cwclock record ls --max 10`
- `delete record abc123` -> `cwclock record delete -i abc123`
- `run job abc123 now` -> `cwclock job run -i abc123`
- `create an export job named "Weekly summary" every monday at 9am emailed to a@b.com for clients c1 and c2` -> `cwclock job create --name "Weekly summary" --cron "0 9 * * 1" --report-types summary-pdf --time-period now()-7d --to a@b.com --client c1 --client c2`
- `preview an invoice for client Acme from Jan 15 9am to Jan 15 noon 2024` (client given by name) -> `cwclock invoice preview --client Acme --begin 2024-01-15T09:00:00 --end 2024-01-15T12:00:00`
- `generate an invoice for client c1 for project p1 only, from now()-30d to now(), save it as invoice.pdf` -> `cwclock invoice generate --client c1 --project p1 --begin now()-30d --end now() -o invoice.pdf`
- `email invoice INV-2024-0007 to its client` (invoice given by number instead of id) -> `cwclock invoice send -i INV-2024-0007`
- `reupload invoice inv123 to external connections` -> `cwclock invoice upload -i inv123`
- `delete invoice inv123` -> `cwclock invoice delete -i inv123`
- `export a summary report as pdf for the last 30 days` -> `cwclock export --begin now()-30d --end now()`
- `export a detailed csv report for client c1 from now()-7d to now()` -> `cwclock export --type detailed --file-format csv --client c1 --begin now()-7d --end now()`
