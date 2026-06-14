# API Specification
## HROS Admin — Super Admin Portal
**Document ID:** HROS-API-001 | **Version:** 1.0
**Standard:** OpenAPI 3.0 (narrative format)
**Status:** Baseline | **Classification:** Internal — Confidential
> `[A]` = Assumption inferred from SaaS HRMS best practice where design was silent

---

## 1. API Foundation

### 1.1 Base URL
```
Production:  https://admin-api.hros.com/api/v1
Staging:     https://staging-admin-api.hros.com/api/v1
Development: http://localhost:4000/api/v1
```

### 1.2 Authentication
All endpoints (except Auth endpoints) require a valid JWT access token:
```
Authorization: Bearer <access_token>
```
- Access tokens expire after **15 minutes** (RS256, as per SRS NFR-S-03)
- Refresh tokens expire after **30 days** (persistent) or at browser session end
- Missing / expired token → `401 Unauthorized`
- Valid token but insufficient permission → `403 Forbidden`

### 1.3 Global Request Headers

| Header | Required | Value |
|--------|----------|-------|
| `Authorization` | Yes (all non-auth routes) | `Bearer <access_token>` |
| `Content-Type` | Yes (POST/PUT/PATCH) | `application/json` |
| `X-Request-ID` | Recommended | UUID v4 for tracing `[A]` |
| `X-CSRF-Token` | Yes (state-changing) | CSRF token from session (NFR-S-11) |

### 1.4 Standard Response Envelopes

**Success (list):**
```json
{
  "data": [ ...items ],
  "meta": {
    "page": 1,
    "per_page": 10,
    "total": 128,
    "total_pages": 13
  }
}
```

**Success (single):**
```json
{
  "data": { ...item }
}
```

**Error:**
```json
{
  "error": {
    "code": "TENANT_CODE_EXISTS",
    "message": "Tenant code already exists.",
    "details": [
      { "field": "code", "issue": "Must be globally unique" }
    ]
  }
}
```

### 1.5 Date/Time Format
All dates and datetimes use **ISO 8601 UTC**: `2024-01-15T10:30:00Z`
Date-only fields use: `2024-01-15`

### 1.6 HTTP Status Codes Used

| Code | Meaning | When Used |
|------|---------|-----------|
| 200 | OK | Successful GET, PUT, PATCH |
| 201 | Created | Successful POST creating a resource |
| 204 | No Content | Successful DELETE |
| 400 | Bad Request | Validation failure, malformed body |
| 401 | Unauthorized | Missing or expired token |
| 403 | Forbidden | Valid token, insufficient permission |
| 404 | Not Found | Resource does not exist |
| 409 | Conflict | Uniqueness violation (e.g., duplicate code) |
| 422 | Unprocessable Entity | Semantic validation failure (e.g., downgrade blocked by usage) |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Unexpected server failure |

### 1.7 Pagination Query Parameters
All list endpoints accept:
```
?page=1&per_page=10&sort_by=created_at&sort_dir=desc
```

### 1.8 Rate Limiting
- `1,000 requests/min` per authenticated admin session (NFR-SC-04)
- Rate limit headers returned on every response: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`

---

## 2. Authentication Endpoints

### 2.1 POST /auth/login
Login with email and password.

**Permission:** Public (no token required)

**Request Body:**
```json
{
  "email": "alex.rivera@hros.com",
  "password": "SecurePass123!",
  "remember_me": true
}
```

| Field | Type | Required | Validation |
|-------|------|----------|-----------|
| email | string | Yes | RFC 5322 format |
| password | string | Yes | Non-empty |
| remember_me | boolean | No | Default: false |

**Success Response — 200 OK** (password-only admin):
```json
{
  "data": {
    "access_token": "eyJhbGciOiJSUzI1NiJ9...",
    "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2g...",
    "token_type": "Bearer",
    "expires_in": 900,
    "admin": {
      "id": "a1b2c3d4-...",
      "name": "Alex Rivera",
      "email": "alex.rivera@hros.com",
      "role": "Super Admin",
      "avatar_initials": "AR"
    }
  }
}
```

**Success Response — 200 OK** (MFA required for Super Admin):
```json
{
  "data": {
    "mfa_required": true,
    "mfa_token": "mfa_sess_abc123",
    "mfa_methods": ["totp", "webauthn"]
  }
}
```

**Error Responses:**
- `401` — `INVALID_CREDENTIALS`: "Invalid email or password"
- `401` — `ACCOUNT_LOCKED`: "Account locked until 2024-01-15T10:30:00Z"
- `403` — `ACCOUNT_INACTIVE`: "Account is disabled. Contact IT Support."

---

### 2.2 POST /auth/mfa/verify
Complete MFA step for Super Admin after successful password auth.

**Permission:** Requires `mfa_token` from `/auth/login`

**Request Body:**
```json
{
  "mfa_token": "mfa_sess_abc123",
  "method": "totp",
  "code": "123456"
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| mfa_token | string | Yes | Short-lived MFA session token from login step |
| method | string | Yes | `totp` or `webauthn` |
| code | string | Cond. | Required for TOTP |

**Success Response — 200 OK:**
```json
{
  "data": {
    "access_token": "eyJhbGciOiJSUzI1NiJ9...",
    "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2g...",
    "token_type": "Bearer",
    "expires_in": 900,
    "admin": { ... }
  }
}
```

**Error Responses:**
- `401` — `MFA_INVALID`: "MFA verification failed"
- `401` — `MFA_TOKEN_EXPIRED`: "MFA session expired. Please log in again."

---

### 2.3 POST /auth/refresh
Issue a new access token using a valid refresh token.

**Permission:** Public (refresh token in body)

**Request Body:**
```json
{
  "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2g..."
}
```

**Success Response — 200 OK:**
```json
{
  "data": {
    "access_token": "eyJhbGciOiJSUzI1NiJ9...",
    "expires_in": 900
  }
}
```

**Error Responses:**
- `401` — `REFRESH_TOKEN_INVALID`: Redirect to login

---

### 2.4 DELETE /auth/session
Logout — invalidate current session tokens.

**Permission:** Authenticated

**Request Body:** None

**Success Response — 204 No Content**

---

### 2.5 POST /auth/password-reset/request
Initiate password reset; always returns success to prevent email enumeration.

**Permission:** Public

**Request Body:**
```json
{
  "email": "alex.rivera@hros.com"
}
```

**Success Response — 200 OK:**
```json
{
  "data": {
    "message": "If an account exists for that email, a reset link has been sent."
  }
}
```

---

### 2.6 POST /auth/password-reset/confirm
Complete password reset using token from email.

**Permission:** Public

**Request Body:**
```json
{
  "token": "uuid-reset-token-here",
  "password": "NewSecurePass123!",
  "password_confirmation": "NewSecurePass123!"
}
```

| Field | Validation |
|-------|-----------|
| token | Must exist, not expired, not used |
| password | Min 10 chars, 1 uppercase, 1 number, 1 special char (AC-AUTH-04) `[A]` |
| password_confirmation | Must match password |

**Success Response — 200 OK:**
```json
{
  "data": {
    "message": "Password updated successfully."
  }
}
```

**Error Responses:**
- `400` — `TOKEN_EXPIRED`: "This link has expired. Request a new password reset."
- `400` — `TOKEN_USED`: "This link has already been used."
- `422` — `PASSWORD_WEAK`: Password does not meet requirements

---

### 2.7 POST /auth/accept-invite
Activate admin account via invitation link.

**Permission:** Public

**Request Body:**
```json
{
  "token": "uuid-invite-token-here",
  "password": "NewSecurePass123!",
  "password_confirmation": "NewSecurePass123!"
}
```

**Success Response — 200 OK:**
```json
{
  "data": {
    "access_token": "...",
    "refresh_token": "...",
    "admin": { ... }
  }
}
```

**Error Responses:**
- `400` — `INVITE_EXPIRED`: "Invitation link expired. Contact your administrator."
- `400` — `INVITE_USED`: "Invitation link already used."

---

## 3. Dashboard Endpoints

### 3.1 GET /dashboard/kpis
Fetch all 5 KPI cards with trend data.

**Permission:** `dashboard.can_view`

**Query Parameters:** None

**Success Response — 200 OK:**
```json
{
  "data": {
    "total_tenants": {
      "value": 1248,
      "trend_pct": 12.0,
      "trend_direction": "up"
    },
    "active_tenants": {
      "value": 1192,
      "stability_label": "Stable"
    },
    "expired_subscriptions": {
      "value": 24,
      "trend_pct": -4.0,
      "trend_direction": "down"
    },
    "total_admin_users": {
      "value": 5842
    },
    "active_plans": {
      "value": 14
    },
    "cached_at": "2024-01-15T10:29:00Z",
    "cache_ttl_seconds": 60
  }
}
```

---

### 3.2 GET /dashboard/activity
Fetch the 20 most recent audit events for the activity feed.

**Permission:** `dashboard.can_view`

**Success Response — 200 OK:**
```json
{
  "data": [
    {
      "id": "log_uuid",
      "tenant_name": "Global Logistics Inc.",
      "tenant_code": "GL-8842",
      "tenant_avatar_initials": "GL",
      "action": "Subscription Renewed",
      "status": "success",
      "created_at": "2023-10-24T14:22:00Z",
      "operator_name": "Auto-System",
      "operator_type": "system"
    }
  ]
}
```

---

### 3.3 GET /dashboard/subscription-trends
Fetch monthly subscription status counts for trend chart.

**Permission:** `dashboard.can_view`

**Query Parameters:**
```
?period=6m   (or 12m)
```

**Success Response — 200 OK:**
```json
{
  "data": {
    "period": "6m",
    "months": [
      {
        "month": "2024-01",
        "active": 1140,
        "trial": 48,
        "expired": 22,
        "total": 1210
      }
    ]
  }
}
```

---

### 3.4 GET /dashboard/export
Export dashboard KPI summary.

**Permission:** `dashboard.can_export`

**Query Parameters:**
```
?format=pdf   (or csv)
```

**Success Response — 200 OK:**
- Content-Type: `application/pdf` or `text/csv`
- Content-Disposition: `attachment; filename="dashboard-export-2024-01-15.pdf"`

---

## 4. Tenant Endpoints

### 4.1 GET /tenants
List all tenants with filters, sort, and pagination.

**Permission:** `tenant_management.can_view`

**Query Parameters:**
```
?page=1
&per_page=10
&status=active,suspended          (comma-separated, multi-select)
&plan_type=enterprise             (plan tier label)
&created_from=2024-01-01
&created_to=2024-12-31
&sort_by=created_at               (name|code|status|plan|created_at|end_date)
&sort_dir=desc                    (asc|desc)
&q=acme                           (global search across name/code/owner email)
```

**Success Response — 200 OK:**
```json
{
  "data": [
    {
      "id": "t_uuid",
      "name": "Global Nexus Corp",
      "code": "GN-8842",
      "avatar_initials": "GN",
      "owner_email": "sarah.j@globalnexus.com",
      "plan": {
        "id": "plan_uuid",
        "name": "Enterprise",
        "tier_label": "UNLIMITED"
      },
      "status": "active",
      "created_at": "2023-10-12T00:00:00Z",
      "end_date": "2025-10-11"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 10,
    "total": 128,
    "total_pages": 13
  },
  "aggregate": {
    "active_tenants": 112,
    "total_mrr": 248500.00,
    "suspended": 9,
    "expiring_30d": 14
  }
}
```

---

### 4.2 POST /tenants
Provision a new tenant. Atomic: creates tenant, owner, subscription, initial admin.

**Permission:** `tenant_management.can_create`

**Request Body:**
```json
{
  "tenant": {
    "name": "Acme Corp Global",
    "code": "ACME_001",
    "legal_name": "Acme Corporation Pty Ltd",
    "industry": "technology",
    "country": "US",
    "timezone": "America/New_York",
    "status": "active"
  },
  "owner": {
    "name": "John Smith",
    "email": "john.smith@acme.com",
    "phone": "+1-555-000-0001"
  },
  "initial_admin": {
    "name": "Jane Admin",
    "email": "jane.admin@acme.com",
    "force_password_reset": true
  },
  "subscription": {
    "plan_id": "plan_uuid_here",
    "billing_cycle": "monthly",
    "start_date": "2024-01-15",
    "end_date": null,
    "trial_enabled": false,
    "trial_end_date": null
  },
  "settings": {
    "default_language": "en-US",
    "default_currency": "USD",
    "employee_limit_override": null,
    "admin_limit_override": null,
    "deployment_notes": "Custom SLA — 4hr support response"
  }
}
```

**Field Validations:**

| Field | Rule |
|-------|------|
| tenant.name | Required, max 200 chars |
| tenant.code | Required, unique, alphanumeric + underscore, max 50 chars |
| tenant.country | ISO 3166-1 alpha-2 |
| tenant.status | `active` \| `pending` \| `suspended` |
| owner.email | Valid email, unique in tenant_owners |
| initial_admin.email | Valid email, unique in admin_users |
| subscription.plan_id | Must reference existing active plan |
| subscription.billing_cycle | `monthly` \| `yearly` |
| subscription.trial_end_date | Required if trial_enabled=true; must be > start_date |
| settings.employee_limit_override | If set, must be ≥ plan.max_employee_limit |

**Success Response — 201 Created:**
```json
{
  "data": {
    "id": "t_uuid_new",
    "name": "Acme Corp Global",
    "code": "ACME_001",
    "status": "active",
    "subscription": {
      "id": "sub_uuid",
      "status": "active",
      "plan": { "id": "...", "name": "Pro" }
    },
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**Error Responses:**
- `409` — `TENANT_CODE_EXISTS`: "Tenant code already exists"
- `409` — `EMAIL_DUPLICATE`: "Email already associated with an admin account"
- `422` — `TRIAL_DATE_INVALID`: "Trial end date must be after start date"
- `422` — `PLAN_NOT_ACTIVE`: "Selected plan is not active"

---

### 4.3 GET /tenants/:id
Fetch a single tenant with all details.

**Permission:** `tenant_management.can_view`

**Success Response — 200 OK:**
```json
{
  "data": {
    "id": "t_uuid",
    "name": "Global Logistics Inc.",
    "code": "GL-8842",
    "legal_name": null,
    "industry": "Logistics",
    "country": "US",
    "timezone": "America/New_York",
    "status": "active",
    "archived_at": null,
    "archive_reason": null,
    "created_by": { "id": "admin_uuid", "name": "Alex Rivera" },
    "created_at": "2023-10-12T00:00:00Z",
    "updated_at": "2024-01-10T08:00:00Z",
    "owner": {
      "id": "owner_uuid",
      "name": "Sarah Johnson",
      "email": "sarah.j@globalnexus.com",
      "phone": "+1-555-000-0042"
    },
    "subscription": {
      "id": "sub_uuid",
      "status": "active",
      "plan": { "id": "plan_uuid", "name": "Enterprise" },
      "billing_cycle": "yearly",
      "next_renewal_date": "2025-10-12"
    }
  }
}
```

**Error Responses:**
- `404` — `TENANT_NOT_FOUND`

---

### 4.4 PUT /tenants/:id
Update an existing tenant. Diff-based: only changed fields written.

**Permission:** `tenant_management.can_update`

**Request Body:** Same structure as POST /tenants — only include fields to change. `code` field is silently ignored (immutable).

**Success Response — 200 OK:**
```json
{
  "data": {
    "id": "t_uuid",
    "name": "Global Logistics Inc.",
    "updated_at": "2024-01-15T10:30:00Z",
    "changed_fields": ["industry"]
  }
}
```

**Error Responses:**
- `404` — `TENANT_NOT_FOUND`
- `409` — `EMAIL_DUPLICATE`
- `422` — Validation error per field

---

### 4.5 POST /tenants/:id/archive
Archive a tenant. Super Admin only.

**Permission:** `tenant_management.can_delete` AND role = Super Admin

**Request Body:**
```json
{
  "confirmation_name": "Global Logistics Inc.",
  "reason": "Contract terminated — churned Q4 2024"
}
```

| Field | Validation |
|-------|-----------|
| confirmation_name | Must exactly match tenant.name (case-sensitive) |
| reason | Required, max 500 chars |

**Success Response — 200 OK:**
```json
{
  "data": {
    "id": "t_uuid",
    "status": "archived",
    "archived_at": "2024-01-15T10:30:00Z",
    "data_deletion_scheduled_at": "2024-04-15T10:30:00Z"
  }
}
```

**Error Responses:**
- `400` — `CONFIRMATION_MISMATCH`: "Confirmation name does not match tenant name"
- `403` — `PERMISSION_DENIED`: "Only Super Admins can archive tenants"
- `409` — `ALREADY_ARCHIVED`: "Tenant is already archived"

---

### 4.6 GET /tenants/search
Autocomplete search for tenants.

**Permission:** `tenant_management.can_view`

**Query Parameters:**
```
?q=global&limit=10
```

**Success Response — 200 OK:**
```json
{
  "data": [
    {
      "id": "t_uuid",
      "name": "Global Logistics Inc.",
      "code": "GL-8842",
      "owner_email": "sarah.j@globalnexus.com",
      "status": "active"
    }
  ]
}
```
- Maximum 10 results; response within 300ms (NFR-P-04)

---

## 5. Subscription Endpoints

### 5.1 GET /tenants/:id/subscription
Fetch full subscription detail for a tenant.

**Permission:** `subscription_management.can_view`

**Success Response — 200 OK:**
```json
{
  "data": {
    "id": "sub_uuid",
    "tenant_id": "t_uuid",
    "plan": {
      "id": "plan_uuid",
      "name": "Enterprise Growth Plan",
      "tier_label": "UNLIMITED",
      "description": "Comprehensive HR management suite..."
    },
    "status": "active",
    "billing_cycle": "yearly",
    "monthly_price_snapshot": 1041.58,
    "start_date": "2023-01-12",
    "end_date": null,
    "next_renewal_date": "2024-01-12",
    "billing_period_start": "2023-01-12",
    "billing_period_end": "2024-01-11",
    "auto_renew": true,
    "trial_used": false,
    "trial_start": null,
    "trial_end": null,
    "paused_at": null,
    "pause_expires_at": null,
    "cancelled_at": null,
    "internal_notes": null,
    "payment_method": {
      "type": "card",
      "brand": "visa",
      "last4": "8829",
      "expiry": "08/26"
    },
    "resource_usage": {
      "employees": { "used": 854, "limit": 1000, "pct": 85.4 },
      "admins":    { "used": 12,  "limit": 20,   "pct": 60.0 },
      "storage_gb":{ "used": 46,  "limit": 50,   "pct": 92.0 },
      "api_calls_month": { "used": 1200000, "limit": 5000000, "pct": 24.0 }
    },
    "quota_overrides": {
      "employee_limit": null,
      "admin_limit": null,
      "storage_limit_gb": null,
      "api_calls_month": null
    },
    "usage_updated_at": "2024-01-15T10:25:00Z",
    "invoices": {
      "data": [
        {
          "id": "inv_uuid",
          "invoice_number": "INV-2023-001",
          "amount": 12499.00,
          "currency": "USD",
          "status": "paid",
          "issued_at": "2023-01-12T00:00:00Z",
          "paid_at": "2023-01-12T01:00:00Z",
          "invoice_type": "standard"
        }
      ],
      "meta": { "page": 1, "per_page": 12, "total": 3 }
    }
  }
}
```

---

### 5.2 PUT /tenants/:id/subscription
Save all subscription settings in one call (billing cycle, auto-renew, quotas, notes).

**Permission:** `subscription_management.can_update`

**Request Body:**
```json
{
  "billing_cycle": "yearly",
  "auto_renew": true,
  "quota_overrides": {
    "employee_limit": 1200,
    "admin_limit": 20,
    "storage_limit_gb": 150,
    "api_calls_month": 600000
  },
  "internal_notes": "Custom agreement — discounted rate approved by VP Sales"
}
```

| Field | Rule |
|-------|------|
| billing_cycle | `monthly` \| `yearly`; effective next period |
| quota_overrides.* | If set, must be ≥ plan default for that field |
| internal_notes | Visible only to Super Admin role |

**Success Response — 200 OK:**
```json
{
  "data": {
    "id": "sub_uuid",
    "billing_cycle": "yearly",
    "auto_renew": true,
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

---

### 5.3 POST /tenants/:id/subscription/change-plan
Upgrade or downgrade subscription plan.

**Permission:** `subscription_management.can_update`

**Request Body:**
```json
{
  "new_plan_id": "plan_enterprise_uuid",
  "effective_date": "immediately",
  "direction": "upgrade"
}
```

| Field | Values | Notes |
|-------|--------|-------|
| new_plan_id | UUID of target plan | Must reference active plan |
| effective_date | `immediately` \| `next_billing_period` | — |
| direction | `upgrade` \| `downgrade` | System validates; mismatches return 422 |

**Success Response — 200 OK (upgrade):**
```json
{
  "data": {
    "subscription_id": "sub_uuid",
    "previous_plan": { "id": "...", "name": "Pro" },
    "new_plan": { "id": "...", "name": "Enterprise" },
    "effective_date": "2024-01-15",
    "proration": {
      "charge": 186.67,
      "remaining_days": 28,
      "currency": "USD"
    },
    "invoice_generated": "INV-2024-002"
  }
}
```

**Error Responses:**
- `422` — `DOWNGRADE_EMPLOYEE_LIMIT`: "Current employee count (842) exceeds Pro plan limit (500)"
- `422` — `DOWNGRADE_STORAGE_LIMIT`: "Current storage (92 GB) exceeds Pro plan limit (100 GB)"
- `422` — `PLAN_NOT_FOUND`
- `422` — `SAME_PLAN`: "Tenant is already on this plan"

---

### 5.4 POST /tenants/:id/subscription/trial
Enable or extend trial period.

**Permission:** `subscription_management.can_update`

**Request Body:**
```json
{
  "action": "enable",
  "trial_end_date": "2024-01-29"
}
```

| action | Precondition |
|--------|-------------|
| `enable` | `trial_used = false` |
| `extend` | `status IN ('trial', 'inactive')` |

**Success Response — 200 OK:**
```json
{
  "data": {
    "subscription_id": "sub_uuid",
    "status": "trial",
    "trial_start": "2024-01-15",
    "trial_end": "2024-01-29",
    "trial_used": true
  }
}
```

**Error Responses:**
- `422` — `TRIAL_ALREADY_USED`: "Trial has already been used for this plan"
- `422` — `TRIAL_DATE_EXCEEDS_MAX`: "New trial end exceeds plan maximum trial days"

---

### 5.5 POST /tenants/:id/subscription/pause
Pause an active subscription.

**Permission:** `subscription_management.can_approve`

**Request Body:**
```json
{
  "reason": "Tenant requested pause during office relocation",
  "auto_resume_at": "2024-04-01"
}
```

**Success Response — 200 OK:**
```json
{
  "data": {
    "subscription_id": "sub_uuid",
    "status": "paused",
    "paused_at": "2024-01-15T10:30:00Z",
    "pause_expires_at": "2024-04-15T10:30:00Z"
  }
}
```

**Error Responses:**
- `422` — `ALREADY_PAUSED`
- `422` — `PAUSE_EXCEEDS_MAX`: "Pause duration cannot exceed 90 days"

---

### 5.6 POST /tenants/:id/subscription/cancel
Cancel subscription at end of current billing period.

**Permission:** `subscription_management.can_delete`

**Request Body:**
```json
{
  "reason": "Customer churned — moved to competitor"
}
```

**Success Response — 200 OK:**
```json
{
  "data": {
    "subscription_id": "sub_uuid",
    "status": "cancelled",
    "access_ends_at": "2024-02-11",
    "data_retention_ends_at": "2024-05-11",
    "cancelled_at": "2024-01-15T10:30:00Z"
  }
}
```

---

### 5.7 GET /tenants/:id/invoices
List all invoices for a tenant (paginated).

**Permission:** `subscription_management.can_view`

**Query Parameters:**
```
?page=1&per_page=12
```

**Success Response — 200 OK:**
```json
{
  "data": [
    {
      "id": "inv_uuid",
      "invoice_number": "INV-2023-001",
      "amount": 12499.00,
      "currency": "USD",
      "status": "paid",
      "invoice_type": "standard",
      "billing_period": { "start": "2023-01-12", "end": "2024-01-11" },
      "issued_at": "2023-01-12T00:00:00Z",
      "paid_at": "2023-01-12T01:00:00Z",
      "download_url": "https://s3.../invoices/INV-2023-001.pdf"
    }
  ],
  "meta": { "page": 1, "per_page": 12, "total": 14, "total_pages": 2 }
}
```

---

### 5.8 GET /tenants/:id/invoices/download-all
Bulk download all invoices as a zip archive.

**Permission:** `subscription_management.can_export`

**Success Response — 200 OK:**
- Content-Type: `application/zip`
- Content-Disposition: `attachment; filename="invoices-ACME-2024.zip"`

---

## 6. Plan Endpoints

### 6.1 GET /plans
List all plans (catalog view).

**Permission:** `plan_management.can_view`

**Query Parameters:**
```
?status=active             (active|inactive|archived|all)
&q=enterprise              (search plan name/code)
```

**Success Response — 200 OK:**
```json
{
  "data": [
    {
      "id": "plan_uuid",
      "name": "Pro",
      "code": "PRO_2024_01",
      "description": "Advanced features for growing teams",
      "tier_label": "SCALE",
      "status": "active",
      "is_featured": true,
      "currency": "USD",
      "monthly_price": 199.00,
      "yearly_price": 1990.00,
      "trial_period_days": 14,
      "max_employee_limit": 500,
      "max_admin_limit": 10,
      "max_location_limit": 5,
      "storage_limit_gb": 100,
      "api_rate_limit": 1000,
      "support_level": "premium",
      "audit_log_retention_days": 365,
      "allow_custom_branding": false,
      "enable_bulk_export": true,
      "features": [
        { "feature_key": "employee_mgmt", "is_enabled": true },
        { "feature_key": "attendance",    "is_enabled": true },
        { "feature_key": "payroll",       "is_enabled": true },
        { "feature_key": "leave_management", "is_enabled": true },
        { "feature_key": "reports",       "is_enabled": true },
        { "feature_key": "api_access",    "is_enabled": true },
        { "feature_key": "expense_mgmt",  "is_enabled": false },
        { "feature_key": "asset_management","is_enabled": false }
      ],
      "active_subscriber_count": 50,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 14 }
}
```

---

### 6.2 POST /plans
Create a new subscription plan.

**Permission:** `plan_management.can_create`

**Request Body:**
```json
{
  "name": "Enterprise Plus",
  "code": "EP_2024_01",
  "description": "Full-featured tier for large enterprises",
  "tier_label": "UNLIMITED",
  "status": "active",
  "is_featured": false,
  "currency": "USD",
  "monthly_price": 499.00,
  "yearly_price": 4990.00,
  "trial_period_days": 14,
  "max_employee_limit": null,
  "max_admin_limit": null,
  "max_location_limit": null,
  "storage_limit_gb": 1024,
  "api_rate_limit": 5000,
  "support_level": "enterprise",
  "audit_log_retention_days": 365,
  "allow_custom_branding": true,
  "enable_bulk_export": true,
  "features": [
    { "feature_key": "employee_mgmt",    "is_enabled": true },
    { "feature_key": "attendance",        "is_enabled": true },
    { "feature_key": "leave_management",  "is_enabled": true },
    { "feature_key": "payroll",           "is_enabled": true },
    { "feature_key": "expense_mgmt",      "is_enabled": true },
    { "feature_key": "asset_management",  "is_enabled": true },
    { "feature_key": "reports",           "is_enabled": true },
    { "feature_key": "api_access",        "is_enabled": true }
  ]
}
```

**Field Validations:**

| Field | Rule |
|-------|------|
| name | Required, max 100 chars |
| code | Required, unique, max 50 chars |
| monthly_price | Required, ≥ 0.00 |
| yearly_price | Required, ≥ 0.00 |
| trial_period_days | 0–365 |
| status | `active` \| `inactive` |
| features | All 8 feature_key entries expected |

**Success Response — 201 Created:**
```json
{
  "data": {
    "id": "plan_uuid_new",
    "name": "Enterprise Plus",
    "code": "EP_2024_01",
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**Error Responses:**
- `409` — `PLAN_CODE_EXISTS`
- `422` — Validation errors per field

---

### 6.3 GET /plans/:id
Fetch a single plan with full details.

**Permission:** `plan_management.can_view`

**Success Response — 200 OK:** (same structure as list item)

---

### 6.4 PUT /plans/:id
Update an existing plan. `code` field is ignored (immutable).

**Permission:** `plan_management.can_update`

**Request Body:** Partial fields accepted (PATCH semantics on PUT).

**Success Response — 200 OK:**
```json
{
  "data": {
    "id": "plan_uuid",
    "name": "Pro Plan",
    "updated_at": "2024-01-15T10:30:00Z",
    "impact_warning": {
      "affected_subscriptions": 50,
      "at_risk_tenants": 0,
      "removed_features": []
    }
  }
}
```

**Error Responses:**
- `404` — `PLAN_NOT_FOUND`
- `409` — `PLAN_CODE_EXISTS`

---

### 6.5 POST /plans/:id/archive
Archive a plan. Prevents new subscriptions; existing subscriptions unaffected.

**Permission:** `plan_management.can_delete`

**Request Body:** None

**Success Response — 200 OK:**
```json
{
  "data": {
    "id": "plan_uuid",
    "status": "archived",
    "archived_at": "2024-01-15T10:30:00Z",
    "existing_subscribers": 12
  }
}
```

**Error Responses:**
- `422` — `LAST_ACTIVE_PLAN`: "Cannot archive the only active plan"

---

### 6.6 PATCH /plans/:id/toggle-status
Toggle plan visibility for new signups (active ↔ inactive).

**Permission:** `plan_management.can_update`

**Request Body:**
```json
{
  "status": "inactive"
}
```

**Success Response — 200 OK:**
```json
{
  "data": { "id": "plan_uuid", "status": "inactive" }
}
```

---

## 7. Admin User Endpoints

### 7.1 GET /admins
List all admin accounts.

**Permission:** `admin_management.can_view`

**Query Parameters:**
```
?page=1&per_page=25
&status=active          (active|inactive|pending)
&role_id=uuid
&q=sarah
```

**Success Response — 200 OK:**
```json
{
  "data": [
    {
      "id": "admin_uuid",
      "name": "Elena Martinez",
      "email": "elena.m@hros.admin",
      "avatar_initials": "EM",
      "role": { "id": "role_uuid", "name": "Super Admin" },
      "status": "active",
      "last_login_at": "2023-10-12T14:00:00Z",
      "created_at": "2023-10-12T00:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 25, "total": 24, "total_pages": 1 },
  "kpis": {
    "total_admins": 24,
    "super_admins": 4,
    "active_now": 12
  }
}
```

---

### 7.2 POST /admins/invite
Invite a new administrator.

**Permission:** `admin_management.can_create`

**Request Body:**
```json
{
  "name": "Sarah Johnson",
  "email": "sarah.j@hros.com",
  "role_id": "role_manager_uuid",
  "account_status": "active"
}
```

| Field | Validation |
|-------|-----------|
| name | Required, max 100 chars |
| email | Valid email, unique in admin_users |
| role_id | Must reference existing role |
| account_status | `active` \| `inactive` |

**Business Rule:** If target role is Super Admin, actor must also be Super Admin (returns `403` otherwise).

**Success Response — 201 Created:**
```json
{
  "data": {
    "id": "admin_uuid_new",
    "name": "Sarah Johnson",
    "email": "sarah.j@hros.com",
    "role": { "id": "...", "name": "Manager" },
    "status": "pending",
    "invite_expires_at": "2024-01-17T10:30:00Z"
  }
}
```

**Error Responses:**
- `409` — `EMAIL_EXISTS`: "An admin with this email already exists"
- `403` — `SUPER_ADMIN_INVITE_RESTRICTED`: "Only Super Admins can invite Super Admins"

---

### 7.3 GET /admins/:id
Fetch a single admin account.

**Permission:** `admin_management.can_view`

**Success Response — 200 OK:**
```json
{
  "data": {
    "id": "admin_uuid",
    "name": "Julian Schmidt",
    "email": "j.schmidt@hros.admin",
    "role": { "id": "role_uuid", "name": "Manager" },
    "status": "active",
    "mfa_enabled": true,
    "last_login_at": "2023-10-12T09:00:00Z",
    "invited_by": { "id": "...", "name": "Alex Rivera" },
    "created_at": "2023-11-04T00:00:00Z"
  }
}
```

---

### 7.4 PUT /admins/:id
Update admin details (name, email).

**Permission:** `admin_management.can_update`

**Request Body:**
```json
{
  "name": "Julian Schmidt-Weber",
  "email": "j.schmidt-weber@hros.admin"
}
```

**Success Response — 200 OK:**
```json
{
  "data": { "id": "admin_uuid", "name": "Julian Schmidt-Weber", "updated_at": "..." }
}
```

---

### 7.5 PATCH /admins/:id/role
Change admin role. Invalidates all active sessions for affected admin.

**Permission:** `admin_management.can_update` AND actor is Super Admin

**Request Body:**
```json
{
  "role_id": "role_billing_admin_uuid"
}
```

**Business Rules:**
- Actor cannot change their own role
- Target role = Super Admin requires actor to be Super Admin

**Success Response — 200 OK:**
```json
{
  "data": {
    "id": "admin_uuid",
    "role": { "id": "...", "name": "Billing Admin" },
    "sessions_invalidated": 3
  }
}
```

**Error Responses:**
- `403` — `SELF_ROLE_CHANGE`: "You cannot change your own role"
- `403` — `SUPER_ADMIN_RESTRICTED`

---

### 7.6 PATCH /admins/:id/status
Activate or deactivate an admin account.

**Permission:** `admin_management.can_update`

**Request Body:**
```json
{
  "status": "inactive"
}
```

**Business Rule:** Admin cannot deactivate their own account (returns `403`).

**Success Response — 200 OK:**
```json
{
  "data": { "id": "admin_uuid", "status": "inactive", "sessions_revoked": 2 }
}
```

**Error Responses:**
- `403` — `SELF_DEACTIVATION`: "You cannot modify your own account"

---

### 7.7 POST /admins/:id/password-reset
Trigger password reset for any admin (admin-initiated).

**Permission:** `admin_management.can_update`

**Request Body:** None

**Success Response — 200 OK:**
```json
{
  "data": {
    "message": "Password reset email sent to j.schmidt@hros.admin"
  }
}
```

---

### 7.8 DELETE /admins/:id
Remove an admin account permanently.

**Permission:** `admin_management.can_delete`

**Business Rule:** Cannot remove self.

**Success Response — 204 No Content**

**Error Responses:**
- `403` — `SELF_REMOVAL`

---

### 7.9 POST /admins/:id/invite/resend
Resend invitation to a pending admin. `[A]`

**Permission:** `admin_management.can_create`

**Success Response — 200 OK:**
```json
{
  "data": {
    "message": "Invitation resent to sarah.j@hros.com",
    "invite_expires_at": "2024-01-17T10:30:00Z"
  }
}
```

---

## 8. Policy Endpoints

### 8.1 GET /policies/roles
List all roles with their permission matrix.

**Permission:** `policy_management.can_view`

**Success Response — 200 OK:**
```json
{
  "data": [
    {
      "id": "role_uuid",
      "name": "Super Admin",
      "description": "Full system access & control",
      "is_system_role": true,
      "permissions": {
        "tenant_management":      { "can_view": true,  "can_create": true,  "can_update": true,  "can_delete": true,  "can_approve": true,  "can_export": true },
        "plan_management":        { "can_view": true,  "can_create": true,  "can_update": true,  "can_delete": true,  "can_approve": true,  "can_export": true },
        "admin_management":       { "can_view": true,  "can_create": true,  "can_update": true,  "can_delete": true,  "can_approve": true,  "can_export": true },
        "policy_management":      { "can_view": true,  "can_create": true,  "can_update": true,  "can_delete": true,  "can_approve": true,  "can_export": true },
        "subscription_management":{ "can_view": true,  "can_create": true,  "can_update": true,  "can_delete": true,  "can_approve": true,  "can_export": true },
        "audit_logs":             { "can_view": true,  "can_create": false, "can_update": false, "can_delete": false, "can_approve": false, "can_export": true },
        "dashboard":              { "can_view": true,  "can_create": false, "can_update": false, "can_delete": false, "can_approve": false, "can_export": true }
      }
    }
  ]
}
```

---

### 8.2 PUT /policies/roles/:role_id/permissions
Save updated permission matrix for a role.

**Permission:** `policy_management.can_update`

**Constraint:** Super Admin role permissions cannot be modified.

**Request Body:**
```json
{
  "permissions": {
    "tenant_management":      { "can_view": true,  "can_create": true,  "can_update": true,  "can_delete": false, "can_approve": false, "can_export": false },
    "plan_management":        { "can_view": true,  "can_create": false, "can_update": false, "can_delete": false, "can_approve": false, "can_export": true },
    "admin_management":       { "can_view": true,  "can_create": false, "can_update": false, "can_delete": false, "can_approve": false, "can_export": false },
    "policy_management":      { "can_view": false, "can_create": false, "can_update": false, "can_delete": false, "can_approve": false, "can_export": false },
    "subscription_management":{ "can_view": true,  "can_create": false, "can_update": true,  "can_delete": false, "can_approve": false, "can_export": false },
    "audit_logs":             { "can_view": true,  "can_create": false, "can_update": false, "can_delete": false, "can_approve": false, "can_export": true },
    "dashboard":              { "can_view": true,  "can_create": false, "can_update": false, "can_delete": false, "can_approve": false, "can_export": false }
  }
}
```

**Success Response — 200 OK:**
```json
{
  "data": {
    "role_id": "role_uuid",
    "saved": true,
    "security_score": {
      "current": 98.2,
      "previous": 97.0,
      "trend": "+1.4%"
    },
    "conflicts": []
  }
}
```

**Error Responses:**
- `403` — `SUPER_ADMIN_IMMUTABLE`: "Super Admin permissions cannot be modified"

---

### 8.3 GET /policies/conditions
List all policy conditions.

**Permission:** `policy_management.can_view`

**Query Parameters:**
```
?role_id=uuid&is_active=true
```

**Success Response — 200 OK:**
```json
{
  "data": [
    {
      "id": "cond_uuid",
      "role": { "id": "...", "name": "Super Admin" },
      "subject": "Super Admin",
      "action": "Create Tenant",
      "condition_expr": null,
      "effect": "ALLOW",
      "label": "System Default",
      "is_active": true,
      "priority": 1,
      "created_by": { "id": "...", "name": "System" }
    },
    {
      "id": "cond_uuid_2",
      "subject": "Billing Admin",
      "action": "Export CSV",
      "condition_expr": "location == 'outside_eu'",
      "effect": "DENY",
      "label": "GDPR Compliance",
      "priority": 5
    }
  ]
}
```

---

### 8.4 POST /policies/conditions
Append a new policy condition.

**Permission:** `policy_management.can_create`

**Request Body:**
```json
{
  "role_id": "role_billing_admin_uuid",
  "subject": "Billing Admin",
  "action": "Export CSV",
  "condition_expr": "location == 'outside_eu'",
  "effect": "DENY",
  "label": "GDPR Compliance",
  "priority": 5
}
```

| Field | Values |
|-------|--------|
| effect | `ALLOW` \| `DENY` |
| priority | Integer; lower = higher priority |

**Success Response — 201 Created:**
```json
{
  "data": {
    "id": "cond_uuid_new",
    "label": "GDPR Compliance",
    "effect": "DENY",
    "is_active": true
  }
}
```

---

### 8.5 PUT /policies/conditions/:id
Update an existing policy condition.

**Permission:** `policy_management.can_update`

**Request Body:** Same as POST

**Success Response — 200 OK:** Updated condition object

---

### 8.6 DELETE /policies/conditions/:id
Remove a policy condition.

**Permission:** `policy_management.can_delete`

**Success Response — 204 No Content**

---

### 8.7 POST /policies/roles
Create a new custom role.

**Permission:** `policy_management.can_create`

**Request Body:**
```json
{
  "name": "Regional Manager",
  "description": "Read-only access for regional offices"
}
```

**Success Response — 201 Created:**
```json
{
  "data": {
    "id": "role_uuid_new",
    "name": "Regional Manager",
    "is_system_role": false,
    "permissions": {}
  }
}
```

**Error Responses:**
- `409` — `ROLE_NAME_EXISTS`

---

### 8.8 POST /policies/health-check
Trigger manual system-wide policy health check.

**Permission:** `policy_management.can_approve`

**Request Body:** None

**Success Response — 200 OK:**
```json
{
  "data": {
    "security_score": 98.2,
    "coverage_pct": 100.0,
    "conflicts": [],
    "warnings": [
      {
        "type": "stale_policy",
        "message": "Policy 'GDPR Compliance' has not been reviewed in 95 days",
        "deduction": -3
      }
    ],
    "checked_at": "2024-01-15T10:30:00Z"
  }
}
```

---

## 9. Audit Log Endpoints

### 9.1 GET /audit-logs
Browse paginated audit log with filters.

**Permission:** `audit_logs.can_view`

**Query Parameters:**
```
?page=1
&per_page=25
&event_type=tenant.created,tenant.archived
&entity_type=tenant
&entity_id=uuid
&operator_id=admin_uuid
&date_from=2024-01-01T00:00:00Z
&date_to=2024-01-31T23:59:59Z
&q=global+logistics
```

**Success Response — 200 OK:**
```json
{
  "data": [
    {
      "id": "log_uuid",
      "event_type": "tenant.created",
      "entity_type": "tenant",
      "entity_id": "t_uuid",
      "entity_name": "Global Logistics Inc.",
      "previous_state": null,
      "new_state": { "name": "Global Logistics Inc.", "status": "active" },
      "diff": null,
      "operator_id": "admin_uuid",
      "operator_name": "Alex Rivera",
      "operator_type": "user",
      "ip_address": "203.0.113.45",
      "user_agent": "Mozilla/5.0 ...",
      "created_at": "2023-10-12T09:30:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 25, "total": 4821, "total_pages": 193 }
}
```

---

### 9.2 GET /audit-logs/export
Export filtered audit log as CSV or PDF.

**Permission:** `audit_logs.can_export`

**Query Parameters:**
```
?format=csv
&event_type=tenant.created
&date_from=2024-01-01T00:00:00Z
&date_to=2024-01-31T23:59:59Z
```

**Success Response — 200 OK:**
- Content-Type: `text/csv` or `application/pdf`
- Content-Disposition: `attachment; filename="audit-log-2024-01-15.csv"`

**Business Rule:** Export action itself is recorded in audit_logs.

---

## 10. Error Code Reference

| Error Code | HTTP | Description |
|-----------|------|-------------|
| `INVALID_CREDENTIALS` | 401 | Email or password incorrect |
| `ACCOUNT_LOCKED` | 401 | Brute-force lockout active |
| `ACCOUNT_INACTIVE` | 403 | Admin account disabled |
| `MFA_INVALID` | 401 | TOTP/WebAuthn code incorrect |
| `MFA_TOKEN_EXPIRED` | 401 | MFA session timed out |
| `TOKEN_EXPIRED` | 400 | Password reset link expired |
| `TOKEN_USED` | 400 | Password reset token already consumed |
| `INVITE_EXPIRED` | 400 | Invitation link expired (>48h) |
| `INVITE_USED` | 400 | Invitation already accepted |
| `PASSWORD_WEAK` | 422 | Does not meet password policy |
| `TENANT_NOT_FOUND` | 404 | Tenant ID does not exist |
| `TENANT_CODE_EXISTS` | 409 | Tenant code already taken |
| `CONFIRMATION_MISMATCH` | 400 | Archive confirmation name wrong |
| `ALREADY_ARCHIVED` | 409 | Tenant already archived |
| `PLAN_NOT_FOUND` | 404 | Plan ID does not exist |
| `PLAN_CODE_EXISTS` | 409 | Plan code already taken |
| `LAST_ACTIVE_PLAN` | 422 | Cannot archive sole active plan |
| `SAME_PLAN` | 422 | Tenant already on target plan |
| `DOWNGRADE_EMPLOYEE_LIMIT` | 422 | Usage exceeds new plan limit |
| `DOWNGRADE_STORAGE_LIMIT` | 422 | Storage exceeds new plan limit |
| `TRIAL_ALREADY_USED` | 422 | Trial already consumed for this plan |
| `TRIAL_DATE_EXCEEDS_MAX` | 422 | Trial end past plan maximum |
| `ALREADY_PAUSED` | 422 | Subscription already paused |
| `PAUSE_EXCEEDS_MAX` | 422 | Pause duration >90 days |
| `EMAIL_EXISTS` | 409 | Admin email already registered |
| `SUPER_ADMIN_INVITE_RESTRICTED` | 403 | Non-SA inviting SA role |
| `SELF_DEACTIVATION` | 403 | Admin modifying own account |
| `SELF_ROLE_CHANGE` | 403 | Admin changing own role |
| `SELF_REMOVAL` | 403 | Admin removing own account |
| `SUPER_ADMIN_IMMUTABLE` | 403 | SA permission matrix is read-only |
| `ROLE_NAME_EXISTS` | 409 | Role name already exists |
| `PERMISSION_DENIED` | 403 | Insufficient RBAC permission |
| `RATE_LIMIT_EXCEEDED` | 429 | >1,000 req/min |

---

*End of API Specification | HROS-API-001 v1.0*