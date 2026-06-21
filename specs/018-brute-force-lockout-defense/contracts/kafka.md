# Kafka Event Contract: email.send.v1

**Feature**: `018-brute-force-lockout-defense`
**Date**: 2026-06-21

This document defines the Kafka message contract published by the admin-service for the lockout email notification event.

---

## Topic

```
email.send.v1
```

**Naming convention**: `<domain>.<event-name>.v<version>` per project coding conventions.

---

## Message Envelope

All messages published to this topic follow the `EventEnvelope[T]` structure:

```json
{
  "id":             "<uuid-v4>",
  "type":           "email.send",
  "source":         "admin-service",
  "version":        1,
  "correlation_id": "<uuid or empty string>",
  "occurred_at":    "<RFC3339 UTC timestamp>",
  "data": {
    "to":       "<recipient email address>",
    "subject":  "<email subject>",
    "template": "<template identifier>",
    "template_data": {
      "<key>": "<value>"
    }
  }
}
```

---

## Lockout Email Payload

When an account is locked due to brute-force protection, the `data` field is populated as follows:

| Field | Type | Example | Description |
|-------|------|---------|-------------|
| `to` | string | `"admin@hros.io"` | Recipient email (also used as Kafka partition key) |
| `subject` | string | `"Account Locked"` | Email subject |
| `template` | string | `"account_locked_notification"` | Template identifier for the notification service |
| `template_data.email` | string | `"admin@hros.io"` | Personalization: recipient's own email |
| `template_data.unlock_at` | string | `"2026-06-21T17:15:00Z"` | RFC3339 unlock timestamp |

---

## Message Key

The Kafka partition key is the recipient email address (`data.to`), encoded as a UTF-8 string. This guarantees that all lockout events for the same user land on the same partition and are processed in order.

---

## Consumer Requirements

Downstream consumers of `email.send.v1` must:

1. Be **idempotent** — the same `id` may be received more than once; consumers must deduplicate by `id`.
2. Handle unknown `template` values gracefully (log and skip).
3. Not assume any specific value for `correlation_id` (may be empty string if context did not carry one).

---

## Producer Reliability

- Published with `sarama.WaitForAll` (all in-sync replicas acknowledge before returning).
- Producer retries up to 5 times on transient errors (inherited from platform Kafka producer config).
- On permanent failure, the error is logged by the admin-service but the account lockout is still applied.
- No outbox / transactional guarantee: at-most-once delivery is the current SLA.
