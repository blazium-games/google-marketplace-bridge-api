# API error codes

Error responses are JSON with a single numeric field `code`:

```json
{"code": 10002}
```

HTTP status codes still reflect the class of error (4xx client, 5xx server). The `code` value identifies the specific condition.

---

## HTTP routing

| Code | HTTP | Meaning |
|------|------|---------|
| 10001 | 405 | The request used an HTTP method other than the one allowed for this path (e.g. not `POST` on `/instantiate`). |

---

## Authentication

| Code | HTTP | Meaning |
|------|------|---------|
| 10002 | 401 | The required security header was missing or its value did not match the configured secret (`SECURITY_HEADER_NAME` / `SECURITY_HEADER_VALUE`). |

---

## JSON payload

| Code | HTTP | Meaning |
|------|------|---------|
| 10101 | 400 | The request body was not valid JSON (syntax error, truncated input, or wrong JSON type for the body). |
| 10102 | 400 | The JSON contained a property name that is not allowed on this endpoint (unknown fields are rejected). |

---

## Instantiate field validation

These codes are returned for `POST /instantiate` when a required string field is missing or only whitespace.

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

## Success responses

Successful calls do not use these codes. For example, `POST /instantiate` returns `201 Created` with a JSON body including `id` and `webhook_notify_at`, not an error `code`.
