# AI instruction 13

## Report filename

The naming convention defined in [ai-instruct-12](./ai-instruct-12.md) is not working because of frontend side:

```javascript
const filename = filenameFromDisposition(response.headers["content-disposition"], `report.${format}`);
```

Fix-it (reuse the backend provided filename).
