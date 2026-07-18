# AI instruction 48

## Bugs on emails templates

### Logo

The default logo in base64 seems broken in the sign up form, I have this:

```html
<td style="padding:24px;text-align:center;background-color:#ffffff;border-bottom:1px solid #e2e8f0">
    <img alt="CWClock" height="36" style="height:36px">
</td>
```

http://localhost:8080/v1/user/confirmation?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3ODQ1NDMxOTYsInB1cnBvc2UiOiJjb25maXJtX2FjY291bnQiLCJzdWIiOiJkY2Y2YjE5ZC0zY2U3LTQ5ZWItYTgwMy01Yzg4YWY4ODMyMGMifQ.ffCTXLgXljxFtD8hdBilHIiarVBGUe38GRqpTHYrRgk

### Body of message

For the invoice email, the body of message seems empty.

Here's the html rendered:

```html
<td style="padding:24px;text-align:center;background-color:#ffffff;border-bottom:1px solid #e2e8f0"> <wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr><wbr></td>
```

There's more `<wbr>` tags I don't know why.

## Mail from

Add a `CWCLOCK_MAIL_FROM` environment variable which is used as expeditor (from) each emails.
This variable should have `noreply@cwcloud.tech` as default value.

## Invoice emails

The `replyTo` must be set to the owner's email of the organization.

## UX/UI

The flag banned should have the same design/color as the current disabled flag (red).
The disabled should become orange.

## Expiration of confirmation's emails

I want the confirmation or forgotten password's emails to expire after 24 hours.

This value should be configurable with the `CWCLOCK_CONFIRMATION_EMAIL_EXPIRATION` environment variable.

The default value should be 24 hours.
