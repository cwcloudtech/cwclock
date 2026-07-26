# cwclock MCP Usage Guide

Use `list_cwclock_commands` or `get_cwclock_command_help` to discover available commands.
Prefer dynamic tools named `cwclock_<command_path>` for direct execution.

## Natural Language To CLI Flags

- `format`, `output format`, `in <format>`, `en <format>` -> `-f <format>` (format can be `json` or `pretty`)
- `filter`, `label filter` -> `--filter <label:value>` (metrics get only)
- `org`, `organization` -> `-o <org-id>` (record commands)
- `client` -> `-c <client-id>` (record --start/--begin)
- `project` -> `-p <project-id>` (record --start/--begin)
- `text`, `description` -> `-t <text>` (record --start/--stop/--begin)
- `record id` -> `-i <record-id>` (record delete)
- `begin`, `from`, `start date` -> `--begin <date>` (or its alias `--from`)
- `end`, `to`, `end date` -> `--end <date>` (or its alias `--to`)

## Important Rules

- Dates for `record --begin`/`--end` (and their `--from`/`--to` aliases) accept ISO-8601, or a
  relative `now()` style expression such as `now()`, `now()-1h`, `now()-1d`.
- `--org`, `--client` and `--project` fall back to the configured `org_id`, `client_id` and
  `project_id` when omitted; if none is available, the command reports which one is missing.

## Examples

- `list all metrics` -> `cwclock metrics ls`
- `get metric cpu_all filtered by instance node-1` -> `cwclock metrics get cpu_all --filter instance:node-1`
- `start a timer for working on the CLI` -> `cwclock record --start -t "working on the CLI"`
- `stop the timer` -> `cwclock record --stop`
- `log time from an hour ago until now for a meeting` -> `cwclock record --begin now()-1h --end now() -t "meeting"`
- `list the last 10 records` -> `cwclock record ls --max 10`
- `delete record abc123` -> `cwclock record delete -i abc123`
