# syntax=docker/dockerfile:1

ARG NODE_IMAGE_TAG=20-alpine
ARG GOLANG_IMAGE_TAG=1.26-alpine
ARG ALPINE_IMAGE_TAG=3.20
ARG NGINX_IMAGE_TAG=1.27-alpine

# Stage ui build
FROM node:${NODE_IMAGE_TAG} AS ui-build
WORKDIR /app
COPY cwclock-ui/package.json cwclock-ui/package-lock.json ./
RUN npm ci
COPY cwclock-ui/ ./
COPY manifest.json ./manifest.json
RUN npm run build:docker

# Stage api build
FROM golang:${GOLANG_IMAGE_TAG} AS api-build
WORKDIR /app
COPY cwclock-api/go.mod cwclock-api/go.sum ./
RUN go mod download
COPY cwclock-api/ ./
COPY manifest.json ./manifest.json
RUN CGO_ENABLED=0 go build -o /out/cwclock-api .

# Stage api run
FROM alpine:${ALPINE_IMAGE_TAG} AS api
RUN apk add --no-cache ca-certificates
COPY --from=api-build /out/cwclock-api /usr/local/bin/cwclock-api
COPY --from=api-build /app/manifest.json /manifest.json
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/cwclock-api"]

# Stage ui run
FROM nginx:${NGINX_IMAGE_TAG} AS ui
COPY --from=ui-build /app/build /usr/share/nginx/html
COPY --from=ui-build /app/manifest.json /usr/share/nginx/html/manifest.json
COPY .docker/nginx/default.conf /etc/nginx/conf.d/default.conf
COPY .docker/nginx/docker-entrypoint.sh /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]

CMD [ "nginx", "-g","daemon off;" ]
