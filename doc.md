# API request reference

Full **error code** list (`10001`–`10901`) is in **`reference.md`**.

| Method & path | Auth | Purpose |
|---------------|------|---------|
| `GET`/`HEAD` `/` | No | Health / liveness (empty **200**). |
| `GET`/`HEAD` `/health` | No | Health / readiness (empty **200**). |
| `POST /instantiate` | Security header | Create record; webhook after `WEBHOOK_NOTIFY_DELAY` (default 5h). |
| `POST /dummy` | Security header | Same body as instantiate; webhook after **5 minutes**; response includes scheduled `time`. |

Both `POST` routes store per-request **`consumer_url`** and **`admin_url`** (random placeholder HTTPS URLs) and send them on the [outbound webhook](#outbound-callback-post-to-webhook_url).

---

## `GET /` and `HEAD /`

| Item | Value |
|------|--------|
| Methods | `GET`, `HEAD` only (others → **405**, code **10001**). |
| Path | Exactly `/` |
| Authentication | None |
| Response | **200 OK** with an empty body. |

---

## `GET /health` and `HEAD /health`

| Item | Value |
|------|--------|
| Methods | `GET`, `HEAD` only (others → **405**, code **10001**). |
| Path | `/health` |
| Authentication | None |
| Response | **200 OK** with an empty body. |

---

## `POST /instantiate`

Creates a row and schedules an outbound `POST` to `webhook_url` after **`WEBHOOK_NOTIFY_DELAY`** (default **5 hours** unless overridden in env).

### HTTP

| Item | Value |
|------|--------|
| Method | `POST` only (else **405** / **10001**). |
| Path | `/instantiate` |
| Request body | JSON object (required). Use `Content-Type: application/json`. |

### Required headers (authentication)

| Header | Description |
|--------|-------------|
| Name | **`SECURITY_HEADER_NAME`** from env (default `X-Instantiate-Secret` if unset). |
| Value | Must equal **`SECURITY_HEADER_VALUE`**. |

Example (names vary by deployment):

```http
GCP_POGR_MARKETPLACE_BRIDGE: 834c7b25-43e2-4494-ad8b-50f2f159c6cf
```

Wrong or missing header → **401** / **10002**.

### JSON body (request)

Required string fields (trimmed); empty/whitespace-only values fail with **10201**–**10206** as in **`reference.md`**.

| Field | JSON key | Description |
|-------|----------|-------------|
| Email | `email` | Contact email. |
| Contract UID | `contract_uid` | Contract identifier. |
| Company | `company` | Company name. |
| Project | `project` | Project name. |
| Webhook URL | `webhook_url` | URL this service will `POST` to (callback). |
| Authorization | `authorization` | Stored and sent as the **`Authorization`** header on that callback. |

**Constraints:** no extra JSON keys (**10102**); valid JSON only (**10101**).

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

| HTTP | Body |
|------|------|
| **201 Created** | `{"success": true}` |

Persisted fields include random **`consumer_url`** and **`admin_url`** for the [outbound payload](#outbound-callback-post-to-webhook_url) (not returned in this response).

### Error responses

JSON only: `{"code": <number>}`. See **`reference.md`**.

---

## `POST /dummy`

Same authentication, JSON body, and error codes as [`POST /instantiate`](#post-instantiate). Shared validation → same **10201**–**10206** / **10101**–**10102** behavior.

| Aspect | `/instantiate` | `/dummy` |
|--------|----------------|----------|
| Webhook delay | `WEBHOOK_NOTIFY_DELAY` (default 5h) | Fixed **5 minutes** after create |
| **201** response body | `{"success": true}` | `{"success": true, "time": "<RFC3339>"}` |

`time` is UTC (RFC 3339) and matches the scheduled **`webhook_notify_at`** for the callback.

Example:

```json
{
  "success": true,
  "time": "2026-04-20T14:35:00Z"
}
```

---

## Outbound callback (`POST` to `webhook_url`)

This service calls **your** `webhook_url` after the row’s scheduled time (`WEBHOOK_NOTIFY_DELAY` for `/instantiate`, **5 minutes** for `/dummy`). It is not a route you call on this API.

| Item | Value |
|------|--------|
| Method | `POST` |
| `Authorization` | Same value you sent in the request body field `authorization`. |
| `Content-Type` | `application/json` |
| Body | See below |

```json
{
  "contract_uid": "ctr-abc-123",
  "consumer_url": "https://consumer.pogr-bridge.invalid/<random-hex>",
  "admin_url": "https://admin.pogr-bridge.invalid/<random-hex>"
}
```

- **`contract_uid`** — From your request.  
- **`consumer_url`** / **`admin_url`** — Generated once per request (distinct URLs); `.invalid` is reserved (placeholders, not resolved by the public DNS).

If the callback returns a non-2xx status, the worker may retry on later polls until a 2xx is received and the delivery is recorded.
