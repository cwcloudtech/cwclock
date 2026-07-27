# AI instruction 89

## Cli

In the `record` command `org` can be passed as an arg as it's already implemented but also a config.

I want also the following commands to be implemented:

```shell
go run main.go organization ls
go run main.go organization create --name <my-org>
go run main.go organization update -i <orga_id> --name <my-org-renamed>
go run main.go organization delete -i <orga_id>
go run main.go organization delete --id <orga_id>
```

Same for client with `--org` param (or current config `organization` if it's set).

Same for project with mandatory `--client` param (with the uuid of the client).

Manage all the mandatory and optional fields properly for those commands.
