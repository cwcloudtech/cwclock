# AI instruction 88

## Cli

### Config

Change the config `endpoint` by `api_url`.

Also the authentication by the header `X-Api-Key` and a config `api_key` (remove all related to `Authorization` headers as config key or username/password basic auth).

Remove also `default_index_name` which is not supposed to be used anymore.

### Clean

Remove the following commands and useless code associate:

```shell
go run main.go details
```

### Record command

#### List

```shell
go run main.go record ls --max 10
```

List the last 10 records (if `--max` is not specified, 10 is the default value).

#### Create

Add the following commands:

```shell
go run main.go record --start
```

Which will start a timer and:

```shell
go run main.go record --stop
```

Which will end the timer and send the time record to cwclock-api with the right endpoint.

Also I want a command:

```shell
go run main.go record --begin <date iso> --end <date iso>
```

Or with this aliases:

```shell
go run main.go record --from <date iso> --to <date iso>
```

I also want this command to inteprete the following syntax

* `now()`
* `now()-1h`
* `now()-1d`
* etc

#### Delete

```shell
go run record delete -i <record_id>
go run record delete --id <record_id>
```

To delete the record.
