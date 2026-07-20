# AI instruction 55

## Contact form

In another java website I have this piece of code:

```java
private static final String X_REAL_IP_HEADER = "X-Real-IP";
private static final String X_FORWARDED_BY_HEADER = "X-Forwarded-By";
private static final String X_CLIENT_IP_HEADER = "X-Client-IP";

String ip = vo.getIp();

if (isNotBlank(ip)) {
    LOGGER.info("[sendContactRequest] ip from VO = {}", ip);
    headers.set(X_CLIENT_IP_HEADER, ip);
    vo.setIp(null);
} else if (httpServletRequest != null) {
    String header = X_REAL_IP_HEADER;
    String realIp = httpServletRequest.getHeader(header);
    if (isBlank(realIp)) {
      header = X_FORWARDED_BY_HEADER;
      realIp = httpServletRequest.getHeader(X_REAL_IP_HEADER);
    }

    LOGGER.info("[sendContactRequest] ip from {} header = {}", header, realIp);
    headers.set(X_CLIENT_IP_HEADER, realIp);
}
```

I want you to do the same:

* Add an optional ip field in your contact form interface contract (which will not be sent in the CWCloud's payload)
* If it's set, fill a `X-Client-IP` header to CWCloud
* If it's not set:
  * Get the value from `X-Real-IP` header if it's set or `X-Forwarded-By` header otherwise and set it in the `X-Client-IP` header for CWCloud

The frontend has nothing to change theoritically.
