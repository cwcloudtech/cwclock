# syntax=docker/dockerfile:1

ARG NODE_IMAGE_TAG=20-alpine
ARG GOLANG_IMAGE_TAG=1.26-alpine
ARG ALPINE_IMAGE_TAG=3.20
ARG NGINX_IMAGE_TAG=1.27-alpine
ARG JDK_IMAGE_TAG=17-jdk-jammy

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
FROM eclipse-temurin:${JDK_IMAGE_TAG} AS mobile-build
ARG GRADLE_VERSION=8.10.2
ARG NODE_MOBILE_MAJOR_VERSION=20
ARG NPM_MOBILE_VERSION=11.19.0
ARG ANDROID_SDK_CMDLINE_TOOLS_VERSION=11076708
ARG ANDROID_PLATFORM=android-35
ARG ANDROID_BUILD_TOOLS=35.0.0

ENV ANDROID_SDK_ROOT=/opt/android-sdk
ENV PATH=${PATH}:/opt/gradle/bin:${ANDROID_SDK_ROOT}/cmdline-tools/latest/bin:${ANDROID_SDK_ROOT}/platform-tools

RUN apt-get update && apt-get install -y --no-install-recommends curl unzip ca-certificates && \
    curl -fsSL https://deb.nodesource.com/setup_${NODE_MOBILE_MAJOR_VERSION}.x | bash - && \
    apt-get install -y nodejs && \
    npm install -g npm@${NPM_MOBILE_VERSION} && \
    curl -fsSL -o /tmp/gradle.zip "https://services.gradle.org/distributions/gradle-${GRADLE_VERSION}-bin.zip" && \
    unzip -q /tmp/gradle.zip -d /opt && mv "/opt/gradle-${GRADLE_VERSION}" /opt/gradle && rm /tmp/gradle.zip && \
    mkdir -p "${ANDROID_SDK_ROOT}/cmdline-tools" && \
    curl -fsSL -o /tmp/cmdline-tools.zip "https://dl.google.com/android/repository/commandlinetools-linux-${ANDROID_SDK_CMDLINE_TOOLS_VERSION}_latest.zip" && \
    unzip -q /tmp/cmdline-tools.zip -d "${ANDROID_SDK_ROOT}/cmdline-tools" && \
    mv "${ANDROID_SDK_ROOT}/cmdline-tools/cmdline-tools" "${ANDROID_SDK_ROOT}/cmdline-tools/latest" && \
    rm /tmp/cmdline-tools.zip && \
    yes | sdkmanager --licenses >/dev/null && \
    sdkmanager "platform-tools" "platforms;${ANDROID_PLATFORM}" "build-tools;${ANDROID_BUILD_TOOLS}" && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY cwclock-mobile/package.json cwclock-mobile/package-lock.json ./
RUN npm ci
COPY cwclock-mobile/ ./
COPY VERSION ./VERSION
RUN VERSION="$(cat VERSION)" && \
    sed -i "s/versionName \"[^\"]*\"/versionName \"${VERSION}\"/" android/app/build.gradle && \
    cd android && gradle assembleRelease --no-daemon && \
    mkdir -p /out && cp app/build/outputs/apk/release/app-release.apk "/out/cwclock-v${VERSION}.apk"

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
COPY --from=mobile-build /out/*.apk /usr/share/nginx/html/
