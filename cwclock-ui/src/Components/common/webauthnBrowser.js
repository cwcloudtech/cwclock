// Bridges the browser's native WebAuthn API (navigator.credentials) with the
// JSON shape the backend's go-webauthn-based MFA endpoints send/expect (see
// ai-instruct-68): every byte field (challenge, credential ids, ...) is a
// base64url string over the wire and an ArrayBuffer/Uint8Array in the
// browser API, so both directions need converting. No WebAuthn JS library is
// used - the wire format matches the spec closely enough that a handful of
// small helpers cover it.

const base64urlToBuffer = (base64url) => {
  const padded = base64url.replace(/-/g, "+").replace(/_/g, "/").padEnd(base64url.length + ((4 - (base64url.length % 4)) % 4), "=");
  const binary = atob(padded);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes.buffer;
};

const bufferToBase64url = (buffer) => {
  const bytes = new Uint8Array(buffer);
  let binary = "";
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
};

// preparePublicKeyCreationOptions converts a CredentialCreation response
// (from the webauthn/register/begin endpoint) into CredentialCreationOptions
// suitable for navigator.credentials.create().
export const preparePublicKeyCreationOptions = ({ publicKey }) => ({
  publicKey: {
    ...publicKey,
    challenge: base64urlToBuffer(publicKey.challenge),
    user: { ...publicKey.user, id: base64urlToBuffer(publicKey.user.id) },
    excludeCredentials: (publicKey.excludeCredentials || []).map((c) => ({ ...c, id: base64urlToBuffer(c.id) })),
  },
});

// preparePublicKeyRequestOptions converts a CredentialAssertion response
// (from a webauthn/begin endpoint) into CredentialRequestOptions suitable
// for navigator.credentials.get().
export const preparePublicKeyRequestOptions = ({ publicKey }) => ({
  publicKey: {
    ...publicKey,
    challenge: base64urlToBuffer(publicKey.challenge),
    allowCredentials: (publicKey.allowCredentials || []).map((c) => ({ ...c, id: base64urlToBuffer(c.id) })),
  },
});

// attestationToJSON serializes a newly created PublicKeyCredential
// (navigator.credentials.create() result) into the JSON shape the
// webauthn/register/finish endpoint expects.
export const attestationToJSON = (credential) => ({
  id: credential.id,
  rawId: bufferToBase64url(credential.rawId),
  type: credential.type,
  response: {
    clientDataJSON: bufferToBase64url(credential.response.clientDataJSON),
    attestationObject: bufferToBase64url(credential.response.attestationObject),
  },
});

// assertionToJSON serializes a login assertion (navigator.credentials.get()
// result) into the JSON shape the webauthn/finish endpoints expect.
export const assertionToJSON = (credential) => ({
  id: credential.id,
  rawId: bufferToBase64url(credential.rawId),
  type: credential.type,
  response: {
    clientDataJSON: bufferToBase64url(credential.response.clientDataJSON),
    authenticatorData: bufferToBase64url(credential.response.authenticatorData),
    signature: bufferToBase64url(credential.response.signature),
    userHandle: credential.response.userHandle ? bufferToBase64url(credential.response.userHandle) : undefined,
  },
});

export const webauthnSupported = () => typeof window !== "undefined" && !!window.PublicKeyCredential;
