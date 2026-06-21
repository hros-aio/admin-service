# API Contracts: Auth Refresh Handler

This document outlines the REST API contracts handled by `AuthHandler` for TSK-AUTH-016.

## 1. Post Login (Updated)

*   **Path**: `POST /v1/auth/login`
*   **Request Type**: `application/json`
*   **Request Payload**:
    ```json
    {
      "email": "admin@example.com",
      "password": "securepassword",
      "remember_me": true
    }
    ```

---

## 2. Post Refresh Token (New Route)

*   **Path**: `POST /v1/auth/refresh`
*   **Request Type**: `application/json`
*   **Request Payload**:
    ```json
    {
      "refresh_token": "valid-refresh-token-string"
    }
    ```
*   **Success Response (200 OK)**:
    ```json
    {
      "access_token": "new-jwt-access-token",
      "refresh_token": "new-refresh-token"
    }
    ```
*   **Error Responses**:
    *   **400 Bad Request**: Missing or empty `refresh_token` string.
    *   **401 Unauthorized**: Invalid or expired refresh token.
    *   **403 Forbidden**: Target admin user is inactive or locked.
    *   **500 Internal Server Error**: Unexpected backend database or cache errors.
