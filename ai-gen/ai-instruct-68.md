# AI instruction 68

## Git external connection

I want an external connection of type git who takes:

* username/password or ssh key to authenticate
* url of the git repo

Same as S3 or drive, read and write in the same folders with the flatten option.

Add the git dependancy in the docker image if it's required but I'd prefer only go library as possible.

The message of commit will be:

> Add invoice {INVOICE_ID} from {period}

or

> Remove invoice {INVOICE_ID}

## MFA

I want the user to be able to register MFA (OTP) app like Google Authenticator scanning QR code or device like yubikey.

Then it's required on the login screen.

An admin can disable MFA for a user.
