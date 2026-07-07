# AI instruction 14

## Frontend UX/UI

All uploadabled images like user's avatar, organization or stamp must be resized to 256x256px and let the user choose the crop of the image.

The `256x256px` must be set an environment variable `IMAGE_SIZE` (and this is the default value).

## Healthcheck endpoint

Add a `GET /v1/health` which answer:

```json
{"status":"ok","alive":true}
```

## Manifest endpoint

Add a `GET /v1/manifest` which answer the content of the `manifest.json` file (it has to be included as an asset during the build and will be replaced dynamically by the cicd pipeline).

## Opentelemetry traces and logs 

I want opentelemetry traces with span on every endpoint call on the backend.

All the logs must also be written in stdout/stderr but also send to opentelemetry with a proper logger and log level.

The opentelemetry target is a single variable `OTEL_ENDPOINT` which can be otlp/grpc or otlp/http with a second variable `OTEL_PROTO`. It will be a single opentelemetry endpoint which will receive metrics, traces and logs.

## Metrics

I want also metrics to be send to the OTEL endpoint like logs and traces but also with a `GET /metrics` endpoint in Prometheus format (it should appear on the openapi page).

It has to contain the default go metrics but also the current metrics:
* counter of each endpoint calls
* average time of each endpoint calls
* counter of users per roles
* counter of clients, projects
* tasks duration in the last 24h (like the summary report stat group by task name and sanitize metrics name replacing spaces with underscores, removing special chars, etc) with id user in the labels and id organization, client and projects

