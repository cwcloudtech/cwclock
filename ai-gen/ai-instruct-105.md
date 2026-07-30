# AI instruction 105

## Mobile app

I want a mobile app which will ask to scan the QR code or the api key (it will parse the api url, api key and default organization inside the qr code).

The mobile app will present a subset of CWClock's feature:

### Time recording

* record start/stop
* record all day with a date picker
* update a record with a mobile screen
* delete a record

### Report generation

Simplified generation screen for summary and detailed reports in PDF format only

### Invoice generation and preview

Implement this.

### Technologies

I want it to be in React Native to be able to re-use some components.
I want it to be in a `cwclock-mobile` folder.

### CICD build

I want the Dockerfile to be completed as it:
* a `mobile-build` stage for Android app wich will reuse the `${VERSION}`
* another stage `ui-and-mobile` which will inherit from `ui` and copy the apk inside `/usr/share/nginx/html`

### Frontend link

I want in the navbar an Android icon to download `https://www.cwclock.me/cwclock-v{version}.apk`, the pattern has to be stored in a static variable `REACT_APP_MOBILE_URL_PATTERN` in the `.env.react` and the version picked from the `manifest.json` file exactly like the printed version.
