# API error codes

Error responses are JSON with a single numeric field `code`:

```json
{"code": 10002}
```

HTTP status codes still reflect the class of error (4xx client, 5xx server). The `code` value identifies the specific condition.

Endpoint reference (requests and success bodies) is in **`doc.md`**.

---

## HTTP routing

| Code | HTTP | Meaning |
|------|------|---------|
| 10001 | 405 | The request used an HTTP method other than allowed for that path. For example: not `POST` on `/instantiate` or `/dummy`; not `GET`/`HEAD` on `/` or `/health`. |

---

## Authentication

| Code | HTTP | Meaning |
|------|------|---------|
| 10002 | 401 | The required security header was missing or its value did not match the configured secret (`SECURITY_HEADER_NAME` / `SECURITY_HEADER_VALUE`). Applies to **`POST /instantiate`** and **`POST /dummy`** (health routes do not use this header). |

---

## JSON payload

| Code | HTTP | Meaning |
|------|------|---------|
| 10101 | 400 | The request body was not valid JSON (syntax error, truncated input, or wrong JSON type for the body). |
| 10102 | 400 | The JSON contained a property name that is not allowed on this endpoint (unknown fields are rejected). |

---

## Instantiate / dummy field validation

These codes are returned for **`POST /instantiate`** and **`POST /dummy`** when a required string field is missing or only whitespace. Validation order is fixed (first failure wins).

| Code | HTTP | Meaning |
|------|------|---------|
| 10201 | 400 | `email` is missing or empty. |
| 10202 | 400 | `contract_uid` is missing or empty. |
| 10203 | 400 | `company` is missing or empty. |
| 10204 | 400 | `project` is missing or empty. |
| 10205 | 400 | `webhook_url` is missing or empty. |
| 10206 | 400 | `authorization` is missing or empty. |

---

## Server / persistence

| Code | HTTP | Meaning |
|------|------|---------|
| 10901 | 500 | The record could not be saved to the database. |

---

## Success responses (no error `code`)

Successful calls do **not** return an error `code`. Examples:

| Situation | HTTP | Body (summary) |
|-----------|------|------------------|
| `GET`/`HEAD` `/` or `/health` | **200** | Empty body. |
| `POST /instantiate` | **201** | `{"success": true}` |
| `POST /dummy` | **201** | `{"success": true, "time": "<RFC3339>"}` — `time` is when the outbound webhook is scheduled (5 minutes after create). |

The outbound **`POST` to your `webhook_url`** is documented in **`doc.md`**; it is not an HTTP response from this API’s public routes, so it does not use these numeric codes.
