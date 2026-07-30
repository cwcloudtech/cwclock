# syntax=docker/dockerfile:1

ARG NODE_IMAGE_TAG=20-alpine
ARG GOLANG_IMAGE_TAG=1.26-alpine
ARG ALPINE_IMAGE_TAG=3.20
ARG NGINX_IMAGE_TAG=1.27-alpine
ARG FLUTTER_IMAGE_TAG=3.44.0-0.3.pre

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

# Stage mobile build (android only)
# Rewritten from React Native to Flutter in ai-instruct-118, after repeated
# RN/Kotlin/Gradle toolchain incompatibilities (ai-instruct-115/116/117) -
# the cirruslabs Flutter image bundles its own compatible JDK/Android SDK/
# Gradle, so none of the manual SDK/Gradle install the RN stage needed is
# necessary anymore.
FROM ghcr.io/cirruslabs/flutter:${FLUTTER_IMAGE_TAG} AS mobile-build
WORKDIR /app
COPY cwclock-mobile/ ./
COPY VERSION ./VERSION
RUN VERSION="$(cat VERSION)" && \
    ANDROID_VERSION_CODE="$(echo "${VERSION}" | awk -F. '{printf "%d%02d%02d", $1, $2, $3}')" && \
    sed -i "s/^version: .*/version: ${VERSION}+${ANDROID_VERSION_CODE}/" pubspec.yaml && \
    flutter pub get && \
    flutter build apk --release && \
    mv build/app/outputs/flutter-apk/app-release.apk "/build/app/outputs/flutter-apk/cwclock-v${VERSION}.apk"

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

# Stage ui-and-mobile run
FROM ui AS ui-and-mobile
COPY --from=mobile-build /app/build/app/outputs/flutter-apk/*.apk /usr/share/nginx/html/
