# Database Domain Model
## HROS Admin — Super Admin Portal
**Document ID:** HROS-DB-001 | **Version:** 1.0
**Database:** PostgreSQL 15+ | **Schema:** hros_admin
> `[A]` = Assumption inferred from SaaS HRMS best practice

---

## 1. Entity Relationship Diagram

```
┌──────────────────┐         ┌────────────────────┐
│   admin_users    │         │      tenants        │
│──────────────────│         │────────────────────│
│ id (PK)          │    ┌───►│ id (PK)             │
│ name             │    │    │ name                │
│ email (UNIQUE)   │    │    │ code (UNIQUE)        │
│ password_hash    │    │    │ legal_name          │
│ role_id (FK)     │    │    │ industry            │
│ status           │    │    │ country             │
│ mfa_enabled      │    │    │ timezone            │
│ mfa_secret       │    │    │ status              │
│ last_login_at    │    │    │ archived_at         │
│ fail_count       │    │    │ archive_reason      │
│ locked_until     │    │    │ created_by (FK)     │◄──┐
│ invited_by (FK)  │◄──┐│    │ created_at          │   │
│ created_at       │   ││    │ updated_at          │   │
│ updated_at       │   ││    └────────────┬────────┘   │
└────────┬─────────┘   ││                 │            │
         │             ││    ┌────────────▼────────┐   │
         │             ││    │   tenant_owners     │   │
         │             ││    │────────────────────│   │
         │             ││    │ id (PK)             │   │
         │             ││    │ tenant_id (FK)      │   │
         │             ││    │ name                │   │
         │             ││    │ email               │   │
         │             ││    │ phone               │   │
         │             ││    │ created_at          │   │
         │             ││    └─────────────────────┘   │
         │             ││                              │
         │             │└──────────────────────────────┘
         │             │
┌────────▼─────────┐   │    ┌────────────────────────┐
│   roles          │   │    │     subscriptions       │
│──────────────────│   │    │────────────────────────│
│ id (PK)          │   │    │ id (PK)                 │
│ name (UNIQUE)    │   │    │ tenant_id (FK)          │◄──┐
│ description      │   │    │ plan_id (FK)            │   │
│ is_system_role   │   │    │ status                  │   │
│ created_at       │   │    │ billing_cycle           │   │
│ updated_at       │   │    │ monthly_price_snapshot  │   │
└──────────────────┘   │    │ start_date              │   │
                        │    │ end_date                │   │
┌──────────────────┐   │    │ next_renewal_date       │   │
│role_permissions  │   │    │ billing_period_start    │   │
│──────────────────│   │    │ billing_period_end      │   │
│ id (PK)          │   │    │ auto_renew              │   │
│ role_id (FK)     │   │    │ trial_used              │   │
│ module           │   │    │ trial_start             │   │
│ can_view         │   │    │ trial_end               │   │
│ can_create       │   │    │ paused_at               │   │
│ can_update       │   │    │ pause_expires_at        │   │
│ can_delete       │   │    │ cancelled_at            │   │
│ can_approve      │   │    │ cancel_reason           │   │
│ can_export       │   │    │ internal_notes          │   │
│ created_at       │   │    │ created_by (FK)         │──►│
│ updated_at       │   │    │ created_at              │
└──────────────────┘   │    │ updated_at              │
                        │    └──────────┬──────────────┘
┌──────────────────┐   │               │
│  policy_conds    │   │    ┌──────────▼──────────────┐
│──────────────────│   │    │   subscription_quotas   │
│ id (PK)          │   │    │────────────────────────│
│ role_id (FK)     │   │    │ id (PK)                 │
│ subject          │   │    │ subscription_id (FK)    │
│ action           │   │    │ employee_limit          │
│ condition_expr   │   │    │ admin_limit             │
│ effect           │   │    │ storage_limit_gb        │
│ label            │   │    │ api_calls_month         │
│ is_active        │   │    │ location_limit          │
│ priority         │   │    │ overridden_by (FK)      │◄─┐
│ created_by (FK)  │──►│    │ overridden_at           │  │
│ created_at       │   │    │ created_at              │  │
│ updated_at       │   │    └─────────────────────────┘  │
└──────────────────┘   │                                  │
                        │    ┌────────────────────────┐   │
┌──────────────────┐   │    │        plans            │   │
│  invite_tokens   │   │    │────────────────────────│   │
│──────────────────│   │    │ id (PK)                 │   │
│ id (PK)          │   │    │ name                    │   │
│ admin_id (FK)    │   │    │ code (UNIQUE)           │   │
│ token (UNIQUE)   │   │    │ description             │   │
│ expires_at       │   │    │ tier_label              │   │
│ used_at          │   │    │ status                  │   │
│ created_by (FK)  │──►│    │ is_featured             │   │
│ created_at       │        │ currency                │   │
└──────────────────┘        │ monthly_price           │   │
                             │ yearly_price            │   │
┌──────────────────┐        │ trial_period_days       │   │
│  session_tokens  │        │ max_employee_limit      │   │
│──────────────────│        │ max_admin_limit         │   │
│ id (PK)          │        │ max_location_limit      │   │
│ admin_id (FK)    │        │ storage_limit_gb        │   │
│ access_token     │        │ api_rate_limit          │   │
│ refresh_token    │        │ support_level           │   │
│ expires_at       │        │ audit_log_retention_days│   │
│ is_persistent    │        │ allow_custom_branding   │   │
│ ip_address       │        │ enable_bulk_export      │   │
│ user_agent       │        │ created_by (FK)         │──►│
│ created_at       │        │ created_at              │
│ revoked_at       │        │ updated_at              │
│ revoke_reason    │        └──────────┬──────────────┘
└──────────────────┘                   │
                             ┌──────────▼──────────────┐
┌──────────────────┐        │   plan_features          │
│   invoices       │        │────────────────────────│
│──────────────────│        │ id (PK)                 │
│ id (PK)          │        │ plan_id (FK)            │
│ invoice_number   │        │ feature_key             │
│ subscription_id  │        │ is_enabled              │
│ tenant_id (FK)   │        │ created_at              │
│ amount           │        └─────────────────────────┘
│ currency         │
│ status           │        ┌─────────────────────────┐
│ billing_period   │        │      audit_logs          │
│ issued_at        │        │────────────────────────│
│ paid_at          │        │ id (PK)                 │
│ refunded_at      │        │ event_type              │
│ payment_method   │        │ entity_type             │
│ stripe_invoice_id│        │ entity_id               │
│ created_at       │        │ entity_name             │
└──────────────────┘        │ previous_state (JSONB)  │
                             │ new_state (JSONB)        │
                             │ diff (JSONB)             │
                             │ operator_id             │
                             │ operator_name           │
                             │ operator_type           │
                             │ ip_address              │
                             │ user_agent              │
                             │ created_at (IMMUTABLE)  │
                             └─────────────────────────┘
```

---

## 2. Table Definitions

### 2.1 admin_users

```sql
CREATE TABLE admin_users (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name            VARCHAR(100) NOT NULL,
  email           VARCHAR(255) NOT NULL UNIQUE,
  password_hash   VARCHAR(255),                    -- NULL for SSO-only accounts
  role_id         UUID NOT NULL REFERENCES roles(id),
  status          VARCHAR(20) NOT NULL DEFAULT 'pending'
                  CHECK (status IN ('active','inactive','pending')),
  mfa_enabled     BOOLEAN NOT NULL DEFAULT FALSE,
  mfa_secret      VARCHAR(255),                    -- encrypted TOTP secret [A]
  last_login_at   TIMESTAMPTZ,
  fail_count      SMALLINT NOT NULL DEFAULT 0,
  locked_until    TIMESTAMPTZ,
  invited_by      UUID REFERENCES admin_users(id), -- NULL for system-created
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_admin_users_email ON admin_users(email);
CREATE INDEX idx_admin_users_role_id ON admin_users(role_id);
CREATE INDEX idx_admin_users_status ON admin_users(status);
```

### 2.2 roles

```sql
CREATE TABLE roles (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name            VARCHAR(50) NOT NULL UNIQUE,
  description     VARCHAR(255),
  is_system_role  BOOLEAN NOT NULL DEFAULT FALSE, -- TRUE = cannot be deleted
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed data
INSERT INTO roles (name, description, is_system_role) VALUES
  ('Super Admin',     'Full system access & control',             TRUE),
  ('Manager',         'Standard operations and employee records', TRUE),
  ('Auditor',         'Read-only reporting and compliance',       TRUE),
  ('Support Lead',    'Handle tickets & tenant details',          TRUE),
  ('Billing Admin',   'Manage plans, invoices & payments',        TRUE),
  ('Content Editor',  'Policy updates & help articles',           TRUE),
  ('Audit Specialist','Read-only logs & exports',                 TRUE);
```

### 2.3 role_permissions

```sql
CREATE TABLE role_permissions (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  role_id     UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  module      VARCHAR(50) NOT NULL
              CHECK (module IN (
                'tenant_management','plan_management',
                'admin_management','policy_management',
                'subscription_management','audit_logs','dashboard'
              )),
  can_view    BOOLEAN NOT NULL DEFAULT FALSE,
  can_create  BOOLEAN NOT NULL DEFAULT FALSE,
  can_update  BOOLEAN NOT NULL DEFAULT FALSE,
  can_delete  BOOLEAN NOT NULL DEFAULT FALSE,
  can_approve BOOLEAN NOT NULL DEFAULT FALSE,
  can_export  BOOLEAN NOT NULL DEFAULT FALSE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (role_id, module)
);
```

### 2.4 policy_conditions

```sql
CREATE TABLE policy_conditions (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  role_id         UUID NOT NULL REFERENCES roles(id),
  subject         VARCHAR(100) NOT NULL,     -- e.g., 'Super Admin', 'Billing Admin'
  action          VARCHAR(100) NOT NULL,     -- e.g., 'Create Tenant', 'Export CSV'
  condition_expr  TEXT,                      -- e.g., "location == 'outside_eu'"
  effect          VARCHAR(5) NOT NULL CHECK (effect IN ('ALLOW','DENY')),
  label           VARCHAR(100),              -- e.g., 'GDPR Compliance'
  is_active       BOOLEAN NOT NULL DEFAULT TRUE,
  priority        SMALLINT NOT NULL DEFAULT 100, -- lower = higher priority
  created_by      UUID NOT NULL REFERENCES admin_users(id),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_policy_cond_role_action ON policy_conditions(role_id, action);
CREATE INDEX idx_policy_cond_active ON policy_conditions(is_active);
```

### 2.5 tenants

```sql
CREATE TABLE tenants (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name            VARCHAR(200) NOT NULL,
  code            VARCHAR(50) NOT NULL UNIQUE,  -- immutable after creation
  legal_name      VARCHAR(200),
  industry        VARCHAR(100),
  country         CHAR(2),                       -- ISO 3166-1 alpha-2
  timezone        VARCHAR(100) NOT NULL DEFAULT 'UTC',
  status          VARCHAR(20) NOT NULL DEFAULT 'pending'
                  CHECK (status IN ('active','pending','suspended','expired','archived')),
  archived_at     TIMESTAMPTZ,
  archive_reason  TEXT,
  created_by      UUID REFERENCES admin_users(id),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tenants_code ON tenants(code);
CREATE INDEX idx_tenants_status ON tenants(status);
CREATE INDEX idx_tenants_created_at ON tenants(created_at DESC);
```

### 2.6 tenant_owners

```sql
CREATE TABLE tenant_owners (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id   UUID NOT NULL UNIQUE REFERENCES tenants(id),
  name        VARCHAR(200) NOT NULL,
  email       VARCHAR(255) NOT NULL,
  phone       VARCHAR(50),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### 2.7 plans

```sql
CREATE TABLE plans (
  id                        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name                      VARCHAR(100) NOT NULL,
  code                      VARCHAR(50) NOT NULL UNIQUE,  -- immutable after creation
  description               TEXT,
  tier_label                VARCHAR(50),       -- e.g., 'STARTER', 'SCALE', 'UNLIMITED'
  status                    VARCHAR(20) NOT NULL DEFAULT 'active'
                            CHECK (status IN ('active','inactive','archived')),
  is_featured               BOOLEAN NOT NULL DEFAULT FALSE,  -- "Most Popular" badge
  currency                  CHAR(3) NOT NULL DEFAULT 'USD',
  monthly_price             NUMERIC(10,2) NOT NULL DEFAULT 0.00,
  yearly_price              NUMERIC(10,2) NOT NULL DEFAULT 0.00,
  trial_period_days         SMALLINT NOT NULL DEFAULT 14,
  max_employee_limit        INTEGER,           -- NULL = unlimited
  max_admin_limit           INTEGER,
  max_location_limit        INTEGER,
  storage_limit_gb          INTEGER,
  api_rate_limit            INTEGER,           -- requests per minute
  support_level             VARCHAR(20) DEFAULT 'basic'
                            CHECK (support_level IN ('basic','premium','enterprise')),
  audit_log_retention_days  INTEGER NOT NULL DEFAULT 365,
  allow_custom_branding     BOOLEAN NOT NULL DEFAULT FALSE,
  enable_bulk_export        BOOLEAN NOT NULL DEFAULT FALSE,
  created_by                UUID REFERENCES admin_users(id),
  created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_plans_status ON plans(status);
CREATE INDEX idx_plans_code ON plans(code);
```

### 2.8 plan_features

```sql
CREATE TABLE plan_features (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  plan_id     UUID NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
  feature_key VARCHAR(50) NOT NULL
              CHECK (feature_key IN (
                'employee_mgmt','attendance','leave_management','payroll',
                'expense_mgmt','asset_management','reports','api_access'
              )),
  is_enabled  BOOLEAN NOT NULL DEFAULT FALSE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (plan_id, feature_key)
);
```

### 2.9 subscriptions

```sql
CREATE TABLE subscriptions (
  id                       UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id                UUID NOT NULL UNIQUE REFERENCES tenants(id),
  plan_id                  UUID NOT NULL REFERENCES plans(id),
  status                   VARCHAR(20) NOT NULL DEFAULT 'pending'
                           CHECK (status IN (
                             'active','trial','inactive','paused',
                             'cancelled','expired','pending'
                           )),
  billing_cycle            VARCHAR(10) NOT NULL DEFAULT 'monthly'
                           CHECK (billing_cycle IN ('monthly','yearly')),
  monthly_price_snapshot   NUMERIC(10,2) NOT NULL, -- price at time of subscription
  start_date               DATE NOT NULL,
  end_date                 DATE,
  next_renewal_date        DATE,
  billing_period_start     DATE,
  billing_period_end       DATE,
  auto_renew               BOOLEAN NOT NULL DEFAULT TRUE,
  trial_used               BOOLEAN NOT NULL DEFAULT FALSE,
  trial_start              DATE,
  trial_end                DATE,
  paused_at                TIMESTAMPTZ,
  pause_expires_at         TIMESTAMPTZ,
  cancelled_at             TIMESTAMPTZ,
  cancel_reason            TEXT,
  internal_notes           TEXT,                   -- Super Admin only
  created_by               UUID REFERENCES admin_users(id),
  created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_tenant_id ON subscriptions(tenant_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_next_renewal ON subscriptions(next_renewal_date);
```

### 2.10 subscription_quotas

```sql
CREATE TABLE subscription_quotas (
  id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  subscription_id   UUID NOT NULL UNIQUE REFERENCES subscriptions(id),
  employee_limit    INTEGER,         -- NULL inherits from plan
  admin_limit       INTEGER,
  storage_limit_gb  INTEGER,
  api_calls_month   INTEGER,
  location_limit    INTEGER,
  overridden_by     UUID REFERENCES admin_users(id),
  overridden_at     TIMESTAMPTZ,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### 2.11 invoices

```sql
CREATE TABLE invoices (
  id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  invoice_number    VARCHAR(50) NOT NULL UNIQUE, -- e.g., INV-2023-001
  subscription_id   UUID NOT NULL REFERENCES subscriptions(id),
  tenant_id         UUID NOT NULL REFERENCES tenants(id),
  amount            NUMERIC(10,2) NOT NULL,
  currency          CHAR(3) NOT NULL DEFAULT 'USD',
  status            VARCHAR(20) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','paid','refunded','failed')),
  billing_period    DATERANGE,
  invoice_type      VARCHAR(20) DEFAULT 'standard'
                    CHECK (invoice_type IN ('standard','proration','trial')),
  issued_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  paid_at           TIMESTAMPTZ,
  refunded_at       TIMESTAMPTZ,
  payment_method    JSONB,                        -- { type, last4, brand, expiry }
  stripe_invoice_id VARCHAR(100),
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invoices_tenant_id ON invoices(tenant_id);
CREATE INDEX idx_invoices_subscription_id ON invoices(subscription_id);
CREATE INDEX idx_invoices_status ON invoices(status);
```

### 2.12 invite_tokens

```sql
CREATE TABLE invite_tokens (
  id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  admin_id     UUID NOT NULL REFERENCES admin_users(id),
  token        VARCHAR(255) NOT NULL UNIQUE,
  expires_at   TIMESTAMPTZ NOT NULL,
  used_at      TIMESTAMPTZ,
  created_by   UUID NOT NULL REFERENCES admin_users(id),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invite_tokens_token ON invite_tokens(token);
CREATE INDEX idx_invite_tokens_admin_id ON invite_tokens(admin_id);
```

### 2.13 session_tokens

```sql
CREATE TABLE session_tokens (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  admin_id        UUID NOT NULL REFERENCES admin_users(id),
  refresh_token   VARCHAR(500) NOT NULL UNIQUE,  -- hashed in storage [A]
  expires_at      TIMESTAMPTZ NOT NULL,
  is_persistent   BOOLEAN NOT NULL DEFAULT FALSE,
  ip_address      INET,
  user_agent      TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  revoked_at      TIMESTAMPTZ,
  revoke_reason   VARCHAR(100)
                  CHECK (revoke_reason IN (
                    'logout','password_change','role_change',
                    'deactivation','expiry','admin_revoke'
                  ))
);

CREATE INDEX idx_session_tokens_admin_id ON session_tokens(admin_id);
CREATE INDEX idx_session_tokens_refresh ON session_tokens(refresh_token);
```

### 2.14 audit_logs (Append-Only)

```sql
CREATE TABLE audit_logs (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  event_type      VARCHAR(100) NOT NULL,   -- e.g., 'tenant.created', 'policy.updated'
  entity_type     VARCHAR(50) NOT NULL
                  CHECK (entity_type IN (
                    'tenant','plan','admin_user','subscription',
                    'policy','invoice','auth'
                  )),
  entity_id       UUID,
  entity_name     VARCHAR(200),
  previous_state  JSONB,
  new_state       JSONB,
  diff            JSONB,                   -- only changed fields
  operator_id     UUID,                    -- NULL for system/bot events
  operator_name   VARCHAR(200),
  operator_type   VARCHAR(20) NOT NULL DEFAULT 'user'
                  CHECK (operator_type IN ('user','system','bot')),
  ip_address      INET,
  user_agent      TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
  -- NO updated_at — this table is append-only
);

-- No FK on operator_id intentionally: even deleted admins' logs must remain
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_event_type ON audit_logs(event_type);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_operator_id ON audit_logs(operator_id);

-- Prevent UPDATE and DELETE at database level
CREATE RULE audit_logs_no_update AS ON UPDATE TO audit_logs DO INSTEAD NOTHING;
CREATE RULE audit_logs_no_delete AS ON DELETE TO audit_logs DO INSTEAD NOTHING;
```

---

## 3. Key Relationships Summary

| Relationship | Type | Description |
|-------------|------|-------------|
| admin_users → roles | Many-to-One | Each admin has one role |
| roles → role_permissions | One-to-Many | Each role has permissions per module |
| roles → policy_conditions | One-to-Many | Each role can have multiple conditions |
| tenants → tenant_owners | One-to-One | Each tenant has one owner record |
| tenants → subscriptions | One-to-One | Each tenant has one active subscription |
| subscriptions → plans | Many-to-One | Many subscriptions can use the same plan |
| subscriptions → subscription_quotas | One-to-One | Optional quota overrides per subscription |
| subscriptions → invoices | One-to-Many | Subscription generates many invoices over time |
| plans → plan_features | One-to-Many | Each plan has feature flags |
| admin_users → invite_tokens | One-to-Many | Admin can have multiple invite tokens (resend) |
| admin_users → session_tokens | One-to-Many | Admin can have multiple active sessions |
| All entities → audit_logs | One-to-Many | All mutations appended to audit log |

---

## 4. Enumerated Values Reference

### Tenant Status
`active` | `pending` | `suspended` | `expired` | `archived`

### Subscription Status
`active` | `trial` | `inactive` | `paused` | `cancelled` | `expired` | `pending`

### Admin User Status
`active` | `inactive` | `pending`

### Invoice Status
`pending` | `paid` | `refunded` | `failed`

### Plan Status
`active` | `inactive` | `archived`

### Policy Effect
`ALLOW` | `DENY`

### Feature Keys
`employee_mgmt` | `attendance` | `leave_management` | `payroll` |
`expense_mgmt` | `asset_management` | `reports` | `api_access`

### Support Levels
`basic` | `premium` | `enterprise`

### Billing Cycle
`monthly` | `yearly`

### Operator Type (Audit)
`user` | `system` | `bot`

---

## 5. Database Constraints & Rules

| Rule | Implementation |
|------|---------------|
| Tenant Code immutable | Application-level: never UPDATE code field; DB trigger `[A]` |
| Plan Code immutable | Same as Tenant Code |
| Audit log append-only | DB RULE prevents UPDATE/DELETE; revoked at app layer too |
| Unique admin email | UNIQUE constraint on admin_users.email |
| One subscription per tenant | UNIQUE constraint on subscriptions.tenant_id |
| One quota row per subscription | UNIQUE constraint on subscription_quotas.subscription_id |
| Session token uniqueness | UNIQUE on session_tokens.refresh_token |
| Quota overrides ≥ plan defaults | Application-level validation before INSERT/UPDATE |

---

## 6. Indexing Strategy

All foreign keys are indexed. Additional indexes:
- `tenants(status)` — frequent filter in list queries
- `tenants(created_at DESC)` — default sort
- `subscriptions(next_renewal_date)` — expiry alerts and MRR jobs
- `subscriptions(status)` — frequent filter
- `audit_logs(entity_type, entity_id)` — per-entity history lookup
- `audit_logs(created_at DESC)` — recent activity feed
- `audit_logs(event_type)` — filtered audit views
- `plans(status)` — catalog filtering

---

*End of Database Domain Model | HROS-DB-001 v1.0*