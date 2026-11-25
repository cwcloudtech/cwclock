# AI instruct 133

## CWAI adapter

Implement an endpoint: 

POST /ai/cwcloud/adapter which implement this interface contract: https://www.cwcloud.tech/docs/tutorials/cwai#external-adapter-interface

```json
{
  "adapter": "test",
  "settings": {
    "max_tokens": 500
  },
  "message": "A prompt",
  "regenerate": false,
  "enable_history": true
}
```

And:

```json
{
  "adapter": "test",
  "settings": {
    "max_tokens": 500
  },
  "messages": [
    {
      "role": "system",
      "message": "You're a tech assistant"
    },
    {
      "role": "user",
      "message": "I need help"
    }
  ],
  "regenerate": false,
  "enable_history": true
}
```

It's important to understand both payload which can be sent by cwcloud.

And reuse AI assistant handler.
Take `adapter` field as the AI subscription targeted (and if it's not matching, take the first subscription).

See the answer's payload in this documentation: https://www.cwcloud.tech/docs/tutorials/cwai#external-adapter-interface
