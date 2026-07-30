// isAndroidDevice reports whether the current browser's User-Agent indicates
// an Android device - used to only show the "download the Android app"
// prompt (navbar, login/signup forms) where the app is actually installable
// (ai-instruct-106).
const isAndroidDevice = () => /Android/i.test(navigator.userAgent || "");

export default isAndroidDevice;
