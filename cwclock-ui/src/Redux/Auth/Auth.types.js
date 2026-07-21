export const register = "auth/register";
export const login = "auth/login";
export const logout = "auth/logout";
export const loading = "auth/loading";
export const error = "auth/error";
export const updatePicture = "auth/updatePicture";
export const updateProfile = "auth/updateProfile";
export const syncUser = "auth/syncUser";
// mfaRequired is dispatched when password login succeeds but the account
// has MFA enabled (see ai-instruct-68): unlike register/login, it must NOT
// persist anything to localStorage - the user isn't authenticated yet until
// the second factor is verified.
export const mfaRequired = "auth/mfaRequired";
