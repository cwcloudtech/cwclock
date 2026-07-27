# AI instruction 97

## Cli

### Record

`--from/--to` or `--begin/--end` should be moved to a sub command `cwclock record create` to keep consistency (with `--text` optional, picking the project's name by default like the frontend).

`stop` shouldn't have mandatory arg and automatically refer to the started record (you can keep infos in a temporary json file inside the `.cwclock` directory containing the config and create if it doesn't exists).
