# AI instruction 90

## Cli

Check the metrics command if it's using the metrics endpoint with the right URL.

Add the following commands:

```shell
go run main.go job create --cron <cron expr>
go run main.go job ls
go run main.go job update -i <export job uuid> # or --id
go run main.go job delete -i <export job uuid> # or --id
```

Manage the mandatory and optional args. For the target I want subcommand like this

```shell
go run main.go job target ls -i <job uuid>
go run main.go job target create -i <job uuid> --to mailto@domain.com
go run main.go job target delete -i <job uuid> --offset 0 # offset in the json array
```

I want the same for external connections in organizations:

```shell
go run main.go organization connection ls -i <orga uuid>
go run main.go organization connection -i <orga uuid>
go run main.go organization connection -i <orga uuid> --offset 0 # offset in the json array
```
