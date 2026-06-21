# API Contracts: Auth Refresh DTOs

This document details the HTTP contract updates for the authentication endpoints.

## Endpoints

### 1. `POST /v1/auth/login`

Updates the existing login endpoint definition to explicitly include `remember_me`.

#### Request Schema (`LoginRequest`)
```json
{
  "email": "admin@hros.com",
  "password": "secure_password",
  "remember_me": true
}
```

#### Responses
- **`200 OK`**: Login successful, returns new JWT and refresh token.
- **`400 Bad Request`**: Validation or syntax error.
- **`401 Unauthorized`**: Invalid credentials.

---

### 2. `POST /v1/auth/refresh`

Exposes the token refresh endpoint to obtain a new token pair using a valid refresh token.

#### Request Schema (`RefreshRequest`)
```json
{
  "refresh_token": "def456..."
}
```

#### Responses
- **`200 OK`**: Token refresh successful, returns new JWT and refresh token.
  ```json
  {
    "access_token": "new_access_jwt...",
    "refresh_token": "new_refresh_token..."
  }
  ```
- **`400 Bad Request`**: Missing or empty `refresh_token` in body.
- **`401 Unauthorized`**: Expired or blacklisted refresh token.
