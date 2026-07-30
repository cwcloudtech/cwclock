// isAndroidDevice reports whether the current browser's User-Agent indicates
// an Android device - used to decide whether the "download the Android app"
// button (navbar, login/signup forms) opens the download link directly, or,
// on devices that can't install the app, opens a QR code modal instead so it
// can be scanned with a phone (ai-instruct-114).
const isAndroidDevice = () => /Android/i.test(navigator.userAgent || "");

export default isAndroidDevice;
