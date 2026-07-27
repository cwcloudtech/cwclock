# AI instruction 99

## Cli

### User

If the uuid is not valid for `cwclock admin user` command with `-i` or `--id` try if it's an email.
Add an endpoint for superuser `GET /admin/users?email=` if it doesn't exists (repeatable `email` which return a list and act with `or` logic). Pick the first element.

### Project / client

When I mention the project and the client like this:

```shell
cwclock record start --project "Platform Engineer" --client "mycli"
```

If no project are matching for the specified client, the cli have to fail fast.
Same for every command with `--project` and `--client`.

Same for `--org` when it's specified and the client is not matching any client in the organization.
