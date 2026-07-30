# AI instruction 118

## Mobile in Flutter

I have enough with the incompatibility between react packages, kotlin, whatever.

That's why I want you to rewrite all `cwcloud-mobile` in Flutter and update the Dockerfile for the build stage.

You can inspire from this:

```shell
ARG FLUTTER_VERSION="3.44.0-0.3.pre"
ARG NGINX_VERSION="alpine"
ARG VERSION="0.0.1"

FROM ghcr.io/cirruslabs/flutter:${FLUTTER_VERSION} AS ui_build

ARG VERSION
ENV VERSION=${VERSION}
ENV CI=true
WORKDIR /app

COPY . .

RUN ANDROID_VERSION_CODE="$(echo "${VERSION}" | awk -F. '{printf "%d%02d%02d", $1, $2, $3}')" && \
    sed -i "s/version: .*/version: ${VERSION}+${ANDROID_VERSION_CODE}/g" pubspec.yaml && \
    flutter pub get && \
    flutter build web && \
    flutter create --platforms=android --project-name cwc_chat_mobile . && \
    mv .docker/AndroidManifest.xml android/app/src/main && \
    mv .docker/gradle.properties android && \
    find android/app/src/main/res -type f -name "ic_launcher*.png" -exec cp assets/images/octopus.png {} \; && \
    flutter build apk

FROM nginx:${NGINX_VERSION} AS ui_run

ARG VERSION
ENV VERSION=${VERSION}

COPY .docker/nginx.conf /etc/nginx/conf.d/default.conf

COPY .docker/docker-entrypoint.sh /docker-entrypoint.sh

COPY --from=ui_build /app/build/web /usr/share/nginx/html

COPY --from=ui_build /app/manifest.json /usr/share/nginx/html/manifest.json

COPY --from=ui_build /app/build/app/outputs/flutter-apk/app-release.apk /usr/share/nginx/html/cwc-chat-v${VERSION}.apk

RUN chmod +x /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]

CMD [ "nginx", "-g","daemon off;" ]
```
