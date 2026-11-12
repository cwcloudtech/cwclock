import { toast } from "react-toastify";
import toastOptions from "../../Redux/toastOptions";

// Copies text and tells the user whether it worked. Every "copy this id"
// control goes through here so the failure path exists in one place:
// navigator.clipboard is absent altogether on a non-secure origin, and
// writeText() rejects whenever the browser withholds the permission. Left
// unhandled - as it was - that rejection surfaces as an uncaught error and
// the user is told nothing at all, while the text silently isn't copied.
const copyToClipboard = (text, t) => {
  if (!navigator.clipboard?.writeText) {
    toast.error(t("common.copyFailed"), toastOptions);
    return Promise.resolve(false);
  }
  return navigator.clipboard
    .writeText(text)
    .then(() => {
      toast.success(t("common.copied"), toastOptions);
      return true;
    })
    .catch(() => {
      toast.error(t("common.copyFailed"), toastOptions);
      return false;
    });
};

export default copyToClipboard;
