# cwclock

Opensource alternative to [Clockify](https://clockify.me) written in Go and react.

## Technologies

- [Go](https://go.dev) and [PostgreSQL](https://www.postgresql.org) for the API
- [React](https://react.dev) for the GUI
- [Flyway](https://flywaydb.org) for the migrations

## Getting started

The whole stack (PostgreSQL, Flyway migrations, the [cwclock-api](./cwclock-api)
backend and the [cwclock-ui](./cwclock-ui) frontend served by nginx) can be
started with a single command:

```shell
docker compose up --build --force-recreate
```

This will:
1. Start a PostgreSQL database.
2. Run the SQL migrations from [cwclock-db](./cwclock-db) with Flyway.
3. Build and start the Go backend, listening on `http://localhost:8080`.
4. Build the React frontend and serve it through nginx on
   `http://localhost:3000`, calling the backend directly at the `API_URL`
   environment variable.

Once it's up, open [`http://localhost:3000`](http://localhost:3000) in your browser.

## Where you can find this repository ?

* Main repo: https://gitlab.cwcloud.tech/oss/cwclock.git
* Github mirror: https://github.com/cwcloudtech/cwclock.git
