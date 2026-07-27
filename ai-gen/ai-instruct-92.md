# AI instruction 92

## Cli

### Jobs

Fix this error:

```shell
$ cwclock job ls
Error: server error (status 404): {"detail":"Not Found"}
```

Job should be the export jobs feature.

### Admin

Add the admin commands following the same model

#### Organizations

```shell
cwclock admin ls
cwclock admin organization delete -i <orga uuid>
cwclock admin organization transfert -i <orga uuid> --owner <new owner uuid>
```

#### User

```shell
cwclock admin user ls
cwclock admin user delete -i <user uuid>
cwclock admin user update -i <user uuid> ## other args
cwclock admin user -i <user uuid> set-role superuser
```
