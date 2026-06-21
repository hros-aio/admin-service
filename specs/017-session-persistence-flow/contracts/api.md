# API Contract Extension: Logout Endpoint

This document outlines the updated request headers supported by the `DELETE /v1/auth/session` route.

## DELETE /v1/auth/session

### Request Headers

The endpoint supports two modes of execution:

#### Mode 1: Access Token Revocation + Session Deletion (New)

Pass the Access Token in the `Authorization` header to blacklist it, and pass the Refresh Token in the `X-Refresh-Token` header to terminate the session database-side:

| Header | Required | Value |
|--------|----------|-------|
| `Authorization` | Yes | `Bearer <access_token>` |
| `X-Refresh-Token` | Yes | `<refresh_token>` |

#### Mode 2: Backward Compatible Session Deletion Only

Pass the Refresh Token in the `Authorization` header. In this mode, no Access Token blacklisting occurs:

| Header | Required | Value |
|--------|----------|-------|
| `Authorization` | Yes | `Bearer <refresh_token>` |
