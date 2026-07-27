# AI instruction 94

## Cli

### External job

Add the following command to run external job:

```shell
cwclock job run -i <job uuid>
```

### Invoice

#### Preview

```shell
cwclock invoice preview --client <client uuid> --begin 2024-01-15T09:00:00 --end 2024-01-15T12:00:00
cwclock invoice preview --client <client uuid> --end 2024-01-15T09:00:00 --to 2024-01-15T12:00:00
```

#### Generate

```shell
cwclock invoice generate --client <client uuid> --begin 2024-01-15T09:00:00 --end 2024-01-15T12:00:00
cwclock invoice generate --client <client uuid> --end 2024-01-15T09:00:00 --to 2024-01-15T12:00:00
```

#### Send

```shell
cwclock invoice send -i <invoice uuid>
cwclock invoice delete -i <invoice uuid>
cwclock invoice upload -i <invoice uuid> # to upload/reupload to external connections
```

### Flag

Always add `--id` equivalent to `-i` everywhere if it's not the case.
