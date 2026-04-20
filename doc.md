# API request reference

This service exposes a single HTTP endpoint. All request bodies are JSON unless noted.

---

## `POST /instantiate`

Creates an instantiation record and schedules a callback to `webhook_url` after the configured delay (default **5 hours**; see `WEBHOOK_NOTIFY_DELAY`).

### HTTP

| Item | Value |
|------|--------|
| Method | `POST` only (any other method returns **405**). |
| Path | `/instantiate` |
| Request body | JSON object (required). `Content-Type` should be `application/json`. |

### Required headers (authentication)

The server compares the following to environment / `.env` values (see `example.env`):

| Header | Description |
|--------|-------------|
| Name | Configured by **`SECURITY_HEADER_NAME`** (default: `X-Instantiate-Secret` if unset). |
| Value | Must exactly match **`SECURITY_HEADER_VALUE`**. |

Example (names depend on your deployment):

```http
GCP_POGR_MARKETPLACE_BRIDGE: 834c7b25-43e2-4494-ad8b-50f2f159c6cf
```

If the header is missing or incorrect, the response is **401** with error code **10002**.

### JSON body (request)

All listed properties are **required** strings. Values are trimmed of leading/trailing whitespace before validation and storage. Empty or whitespace-only values are rejected.

| Field | JSON key | Description |
|-------|------------|-------------|
| Email | `email` | Contact email. |
| Contract UID | `contract_uid` | Identifier for the contract. |
| Company | `company` | Company name. |
| Project | `project` | Project name. |
| Webhook URL | `webhook_url` | HTTPS/HTTP URL the service will `POST` to after the delay (see [Outbound callback](#outbound-callback-post-to-webhook_url)). |
| Authorization | `authorization` | Secret/token stored by the server and sent back as the **`Authorization`** header on the webhook request (so your receiver can verify the callback). |

**Constraints**

- Unknown JSON properties are **not** allowed (`DisallowUnknownFields`). Extra keys produce error code **10102**.
- The body must be a single JSON object; malformed JSON produces **10101**.

Example:

```json
{
  "email": "user@example.com",
  "contract_uid": "ctr-abc-123",
  "company": "Acme Corp",
  "project": "Billing",
  "webhook_url": "https://partner.example.com/hooks/instantiate",
  "authorization": "Bearer your-callback-secret"
}
```

### Success response

| HTTP status | Body |
|-------------|------|
| **201 Created** | JSON object confirming the record was accepted. |

```json
{
  "success": true
}
```

| Field | Type | Description |
|-------|------|-------------|
| `success` | boolean | Always `true` on **201** for this endpoint. |

### Error responses

Errors return JSON **`{"code": <number>}`** only (no message string). HTTP status indicates the class of error (4xx/5xx).

See **`reference.md`** for the full list of codes and meanings (10001–10901).

---

## Outbound callback (`POST` to `webhook_url`)

This is **not** an endpoint you call on this API; it is the request **this service** makes to **your** `webhook_url` after `WEBHOOK_NOTIFY_DELAY`, using the stored row.

| Item | Value |
|------|--------|
| Method | `POST` |
| Header `Authorization` | Value from the original request body field `authorization`. |
| Header `Content-Type` | `application/json` |
| Body | `{"contract_uid":"<stored contract_uid>"}` |

On non-2xx responses, the server may retry on later polls until a 2xx is received and the row is marked delivered.
