# System Requirements Specification (SRS)
## HROS Admin — Super Admin Portal
**Document ID:** HROS-SRS-001 | **Version:** 1.0 | **Standard:** IEEE 830-1998 (adapted)
**Status:** Baseline | **Classification:** Internal — Confidential
> `[A]` = Assumption inferred from SaaS HRMS best practice where design was silent

---

## 1. Introduction

### 1.1 Purpose
This SRS defines all system-level requirements for the HROS Admin Super Admin Portal — a centralized, multi-tenant management console used exclusively by internal HROS operators. It is the authoritative reference for engineering, QA, and architecture teams.

### 1.2 Scope
**In Scope:**
- Secure authentication and session management
- Real-time operational dashboard
- Full tenant lifecycle management (create → update → archive)
- Subscription and billing plan administration
- Role-based access control and policy engine
- Audit logging and compliance reporting

**Out of Scope:**
- Tenant-facing HRMS application and employee portals
- Payroll processing engine
- Payment gateway internals (Stripe handles charge execution)
- Mobile native applications (future enhancement)

### 1.3 Definitions

| Term | Definition |
|------|-----------|
| Tenant | A client organization provisioned on the HROS platform |
| Super Admin | HROS internal operator with unrestricted system access |
| Plan | A subscription tier defining feature access and resource limits |
| MRR | Monthly Recurring Revenue |
| RBAC | Role-Based Access Control |
| MFA | Multi-Factor Authentication |
| SSO | Single Sign-On via federated identity provider |
| Zero-Trust | Security model requiring explicit verification for every action |
| GDPR | General Data Protection Regulation (EU) |
| Proration | Billing adjustment for mid-cycle plan changes |
| Privilege Creep | Accumulation of excessive permissions over time |
| SLA | Service Level Agreement |

### 1.4 System Context Diagram

```
┌─────────────────────────────────────────────────────────┐
│                 HROS Admin Portal (SPA)                  │
│              [Super Admin Control Plane]                  │
│  Auth | Dashboard | Tenants | Plans | Admins | Policies  │
└──────────────────────────┬──────────────────────────────┘
                           │ HTTPS / REST API (TLS 1.3)
           ┌───────────────┼───────────────┐
           ▼               ▼               ▼
    ┌────────────┐  ┌────────────┐  ┌───────────────┐
    │  Auth Svc  │  │ Tenant Svc │  │  Billing Svc  │
    └────────────┘  └────────────┘  └───────────────┘
    ┌────────────┐  ┌────────────┐  ┌───────────────┐
    │ Policy Svc │  │  Plan Svc  │  │   Audit Svc   │
    └────────────┘  └────────────┘  └───────────────┘
           │               │               │
           └───────────────┼───────────────┘
                           ▼
    ┌──────────────────────────────────────────────┐
    │              Data Persistence Layer           │
    │  PostgreSQL 15+ │ Redis 7+ │ S3 Object Store │
    └──────────────────────────────────────────────┘
           │               │               │
    ┌──────┴──────┐ ┌──────┴──────┐ ┌────┴─────────┐
    │  Email/SES  │ │  Stripe API │ │  SSO / IdP   │
    └─────────────┘ └─────────────┘ └──────────────┘
```

---

## 2. Functional Requirements

### 2.1 Authentication & Session Management

#### SRS-AUTH-001: Credential Login
The system SHALL authenticate admin users via email + password.
- Input: email (string), password (string, masked by default)
- Processing: validate format → bcrypt hash compare → issue JWT on match
- Output: access token (15 min) + refresh token (30 days if persistent session selected)
- Failure: generic "Invalid credentials" — no field-specific detail revealed

#### SRS-AUTH-002: Account Lockout
After 5 consecutive failed login attempts within 15 minutes `[A]`, the system SHALL:
- Lock the account for 30 minutes `[A]`
- Email the account holder with lock notification and unlock timestamp
- Require Super Admin intervention for immediate unlock `[A]`

#### SRS-AUTH-003: SSO Authentication
The system SHALL support federated authentication via SAML 2.0 and OIDC.
- System redirects to IdP → receives signed assertion → maps to admin account
- No matching admin account → error: "No admin account linked to this identity"

#### SRS-AUTH-004: Biometric Authentication
The system SHALL support device biometric login via WebAuthn/FIDO2. `[A]`
- Device credential registered on first setup → authenticate via platform authenticator

#### SRS-AUTH-005: MFA Enforcement
The system SHALL enforce MFA for all Super Admin logins. `[A]`
- Supported methods: TOTP (RFC 6238), WebAuthn
- MFA challenge issued after successful credential step; session not created until MFA passes

#### SRS-AUTH-006: Persistent Session
The system SHALL support optional 30-day persistent sessions via "Keep me logged in" checkbox.
- Implemented via long-lived refresh token; access tokens still expire at 15 minutes

#### SRS-AUTH-007: Password Reset
The system SHALL provide self-service password reset via email token.
- Token validity: 60 minutes `[A]`, single-use
- On reset: all existing sessions for that admin invalidated

#### SRS-AUTH-008: Session Invalidation
The system SHALL invalidate all active sessions for an admin upon:
- Explicit logout
- Password change (self or admin-initiated)
- Role assignment change
- Account deactivation

---

### 2.2 Dashboard

#### SRS-DASH-001: KPI Metric Cards
The system SHALL display 5 real-time metric cards on the dashboard:
- Total Tenants (with % change vs. prior 30 days)
- Active Tenants (with stability label)
- Expired Subscriptions (with % change trend)
- Total Admin Users
- Active Plans
- Data cache TTL: 60 seconds `[A]`

#### SRS-DASH-002: Subscription Trend Chart
The system SHALL render a time-series subscription status chart with:
- Toggle between 6-month (default) and 12-month views
- Series: Active, Trial, Expired, Total subscription counts per month
- Hover tooltip showing exact values per period `[A]`

#### SRS-DASH-003: Recent Activity Feed
The system SHALL display the 20 most recent administrative events `[A]` showing:
- Tenant name + avatar initials
- Action description
- Status badge (Success / Applied / Pending / Failed)
- Datetime (UTC, displayed in admin's local timezone)
- Operator identity (user name or system process)
- Feed auto-refreshes every 60 seconds `[A]`

#### SRS-DASH-004: Quick Actions
The system SHALL display 4 quick-action navigation tiles:
Create Tenant | Create Plan | Invite Admin | Manage Policies

#### SRS-DASH-005: Report Export
The system SHALL allow authorized admins to export the dashboard summary as PDF or CSV. `[A]`

---

### 2.3 Tenant Management

#### SRS-TM-001: Tenant List with Filters
The system SHALL display a paginated tenant list (10 per page `[A]`) with:
- Columns: Name, Code, Owner Email, Plan, Status, Created, Expiry
- Filters: Status (multi-select), Plan Type, Created Date range
- Sort: any column ascending/descending; default: Created Date DESC
- Footer: aggregate stats (Active Tenants, Total MRR, Suspended, Expiring 30d)

#### SRS-TM-002: Create Tenant
The system SHALL allow Super Admins to provision a tenant via a 5-section form.
- **Required fields:** Tenant Name, Tenant Code, Owner Name, Owner Email, Admin Name, Admin Email, Select Plan
- **Post-save actions:** create tenant namespace, activate subscription, send onboarding email to Admin Email, write audit log
- **Constraint:** Tenant Code globally unique, immutable after creation

#### SRS-TM-003: Update Tenant
The system SHALL allow authorized admins to update any tenant field except Tenant Code.
- All fields pre-populated from current record
- Diff-based save: only changed fields written; each change logged with old/new values

#### SRS-TM-004: Archive Tenant
The system SHALL allow Super Admins to soft-delete a tenant.
- Requires typed confirmation (tenant name) `[A]`
- Effects: status → Archived, all subscriptions → Cancelled, all tenant admin sessions invalidated
- Data retained for 90 days before scheduled deletion `[A]`
- Permission: Super Admin only

#### SRS-TM-005: Tenant Search
The system SHALL provide global search across tenant name, code, and owner email.
- Autocomplete results within 300 ms, max 10 suggestions

---

### 2.4 Subscription Management

#### SRS-SUB-001: Subscription Detail View
The system SHALL display a full subscription detail page per tenant showing:
- Plan name, billing cycle, start/next renewal dates, amount
- Resource usage meters (Employees, Admins, Storage, API Calls) with percentage bars
- Color coding: ≥90% = red, ≥70% = orange, <70% = blue `[A]`
- Payment method card (last 4 digits, expiry, brand)
- Billing history table (last 12 invoices `[A]`): ID, date, amount, status, download

#### SRS-SUB-002: Plan Upgrade
The system SHALL support mid-cycle plan upgrades with:
- Proration calculation: `(new_daily_rate - old_daily_rate) × remaining_days`
- Effective date option: Immediately or Next Billing Period
- Proration preview shown in confirmation modal before commit

#### SRS-SUB-003: Plan Downgrade
The system SHALL validate usage against new plan limits before allowing downgrade.
- Block if: current employees > new limit, current admins > new limit, or storage used > new limit
- Schedule for end of billing period (no early termination credit `[A]`)

#### SRS-SUB-004: Manual Quota Override
Super Admins SHALL override per-tenant limits (employees, admin seats, storage, API calls).
- Override values must be ≥ plan defaults `[A]`
- Override is tenant-specific and persists until explicitly removed

#### SRS-SUB-005: Trial Management
The system SHALL support trial period enable and extend operations.
- Enable: sets status = Trial, sets trial_end_date
- Extend: updates trial_end_date; cannot exceed plan's max_trial_days `[A]`
- Constraint: one trial per tenant per plan lifetime

#### SRS-SUB-006: Pause / Resume
The system SHALL allow Super Admins to pause a subscription for up to 90 days. `[A]`
- Billing halted; tenant access suspended; data preserved
- Auto-resume OR auto-cancel at max pause duration (configurable per tenant) `[A]`

#### SRS-SUB-007: Cancel Subscription
The system SHALL allow cancellation effective at end of current billing period.
- Requires explicit confirmation dialog with impact summary
- Status → Cancelled; data retention countdown begins at period end

#### SRS-SUB-008: Billing Cycle Toggle
The system SHALL allow switching billing cycle between Monthly and Yearly.
- Effective at next billing period start
- Yearly pricing applies plan's configured yearly_price

#### SRS-SUB-009: Auto-Renew Toggle
The system SHALL allow enabling/disabling auto-renewal per subscription.
- When disabled: expiry warning emails sent at 30d, 7d, 1d before end `[A]`

#### SRS-SUB-010: Internal Admin Notes
Super Admins SHALL add private subscription notes visible only to Super Admin role.

---

### 2.5 Plan Management

#### SRS-PM-001: Plan Catalog View
The system SHALL display all active plans as comparison cards showing:
- Tier label, plan name, monthly price (billed annually)
- Max Employees, Max Admins, Storage limit
- Included features list with checkmarks
- "Most Popular" badge on designated plan
- Module Availability Status matrix at page bottom

#### SRS-PM-002: Create Plan
Super Admins SHALL create plans via a 5-section form:
1. Basic Information (name, code, description, status toggle)
2. Pricing & Billing (currency, monthly price, yearly price, trial days)
3. Capacity & Resource Limits (employees, admins, locations, storage, API rate)
4. Feature Permissions (8 module checkboxes)
5. Advanced Administration (support level, audit log retention, branding, bulk export)
- Plan activates immediately upon creation
- Plan Code: globally unique, immutable after creation

#### SRS-PM-003: Edit Plan
Super Admins SHALL edit plans via full-page form OR inline slide-in panel.
- System displays downstream impact warning before changes are saved
- Inline panel fields: name, monthly price, yearly discount %, max employees, max admins, module toggles

#### SRS-PM-004: Archive Plan
Super Admins SHALL archive plans to prevent new subscriptions.
- Existing tenant subscriptions unaffected
- Cannot archive if it is the only active plan `[A]`

#### SRS-PM-005: Plan Status Toggle
Admins SHALL toggle a plan's visibility for new signups independently of archiving.

---

### 2.6 Admin Management

#### SRS-AM-001: Admin List
The system SHALL display all admin accounts with KPI cards (Total, Super Admins, Active Now) and a table showing name/email, role, status, last login, created date.

#### SRS-AM-002: Invite Administrator
Super Admins SHALL invite new admins by filling: Full Name, Work Email, Administrative Role, Account Status (Active/Inactive).
- System sends email invitation with link valid for 48 hours
- Link leads to password-set page; account activates on completion
- Constraint: only Super Admin can invite other Super Admins

#### SRS-AM-003: Admin Account Actions
Per-admin row SHALL expose: Edit Details, Change Role, Activate/Deactivate, Reset Password, View Activity Log, Remove Admin.
- Self-protection: admin cannot deactivate or remove their own account
- Role changes to/from Super Admin require Super Admin actor

---

### 2.7 Policy Management

#### SRS-POL-001: Role Permission Matrix
The system SHALL display and allow editing of a matrix with:
- Rows: Tenant Mgmt, Plan Mgmt, Admin Mgmt, Policy Mgmt (extensible)
- Columns: View, Create, Update, Delete, Approve, Export
- Global Select buttons: All / None per role
- Super Admin matrix is always full-access and read-only in the UI `[A]`

#### SRS-POL-002: Policy Condition Builder
Super Admins SHALL define policy conditions with: Subject, Action, Condition expression, Effect (Allow/Deny).
- Conditions evaluated at runtime; DENY overrides ALLOW for same action/subject `[A]`
- GDPR-tagged rules evaluated first for any EU data operation `[A]`

#### SRS-POL-003: Conflict Detection
System SHALL automatically detect privilege conflicts after every Save.
- Output: Security Score (0–100), list of conflicts, coverage percentage

#### SRS-POL-004: Custom Role Creation
Super Admins SHALL create custom roles starting with zero permissions.

#### SRS-POL-005: Audit Log from Policy
The "Audit Logs" button in Policies SHALL navigate to a policy-event-filtered audit log view.

---

## 3. Non-Functional Requirements

### 3.1 Performance

| ID | Requirement | Target |
|----|-------------|--------|
| NFR-P-01 | Dashboard initial load | < 2.0s (P95) |
| NFR-P-02 | Tenant list load (10 rows) | < 1.5s (P95) |
| NFR-P-03 | Form submission response | < 3.0s (P99) |
| NFR-P-04 | Search autocomplete | < 300ms (P95) |
| NFR-P-05 | API read endpoints | < 500ms (P99) |
| NFR-P-06 | API write endpoints | < 2.0s (P99) |
| NFR-P-07 | Activity feed latency | ≤ 60s behind real-time |
| NFR-P-08 | Resource usage freshness | ≤ 5 minutes |

### 3.2 Security

| ID | Requirement | Specification |
|----|-------------|--------------|
| NFR-S-01 | Transport encryption | TLS 1.3 preferred; TLS 1.2 minimum |
| NFR-S-02 | Password hashing | bcrypt, cost factor ≥ 12 |
| NFR-S-03 | JWT algorithm | RS256; access 15 min, refresh 30 day |
| NFR-S-04 | MFA | Enforced for Super Admin; optional others `[A]` |
| NFR-S-05 | Brute-force | Lock after 5 failures; 30-min cooldown |
| NFR-S-06 | Data at rest | AES-256 for PII fields `[A]` |
| NFR-S-07 | RBAC | Enforced at API gateway AND service layer |
| NFR-S-08 | GDPR | Location-based export blocking via policy engine |
| NFR-S-09 | Audit immutability | Append-only; no UPDATE/DELETE on audit_logs `[A]` |
| NFR-S-10 | Input sanitization | XSS and SQL injection protection on all inputs |
| NFR-S-11 | CSRF | CSRF tokens on all state-changing requests `[A]` |
| NFR-S-12 | Secrets | No secrets in client-side code or logs `[A]` |

### 3.3 Availability & Reliability

| ID | Requirement | Target |
|----|-------------|--------|
| NFR-A-01 | Uptime SLA | 99.9% monthly (≤ 8.7 hrs downtime/year) |
| NFR-A-02 | Planned maintenance | Off-hours with 48h advance notice `[A]` |
| NFR-A-03 | Daily backup | Full backup + point-in-time recovery `[A]` |
| NFR-A-04 | RTO | < 4 hours `[A]` |
| NFR-A-05 | RPO | < 1 hour `[A]` |

### 3.4 Scalability

| ID | Requirement | Target |
|----|-------------|--------|
| NFR-SC-01 | Tenant records | 10,000+ without degradation `[A]` |
| NFR-SC-02 | Concurrent admin sessions | 500+ `[A]` |
| NFR-SC-03 | Audit log volume | 10M+ entries `[A]` |
| NFR-SC-04 | API throughput | 1,000 req/min per admin session `[A]` |

### 3.5 Compatibility

| Requirement | Target |
|-------------|--------|
| Browsers | Chrome 120+, Firefox 121+, Safari 17+, Edge 120+ |
| Minimum viewport | 1280 × 800 px |
| Runtime (server) | Node.js 20 LTS `[A]` |
| Database | PostgreSQL 15+ `[A]` |
| Cache | Redis 7+ `[A]` |

### 3.6 Usability

- All destructive actions require explicit confirmation dialog
- Inline real-time form validation on field blur
- Unsaved changes trigger sticky indicator + navigation warning
- Breadcrumb on all sub-pages
- Warning banners for sensitive/high-impact operations
- Loading spinners/skeletons for all async operations `[A]`

---

## 4. Interface Requirements

### 4.1 User Interface
- Single-page application (SPA), rendered in browser
- Left fixed sidebar: navigation links with icons
- Fixed top bar: global search, notifications bell, help, user identity
- Main content area: breadcrumb + page title + contextual action buttons
- Slide-in right panels for secondary forms (Invite Admin, Edit Plan)
- Modal dialogs for confirmations and destructive actions

### 4.2 API Interface
- RESTful JSON API, versioned at `/api/v1/`
- Authentication: `Authorization: Bearer <access_token>` header
- Content-Type: `application/json`
- Pagination envelope: `{ data: [...], meta: { page, per_page, total, total_pages } }`
- Error envelope: `{ error: { code, message, details: [...] } }`
- Date format: ISO 8601 UTC (`2024-01-15T10:30:00Z`)

### 4.3 External Service Interfaces

| System | Protocol | Purpose |
|--------|---------|---------|
| Email / AWS SES `[A]` | SMTP / API | Invitations, alerts, onboarding emails |
| Stripe `[A]` | REST API | Subscription billing and invoice generation |
| Identity Provider `[A]` | SAML 2.0 / OIDC | SSO authentication |
| AWS S3 `[A]` | SDK | Export file storage and download links |
| Audit Service | Internal gRPC `[A]` | Append-only event persistence |

---

## 5. Constraints & Assumptions

### 5.1 Constraints
1. Portal accessible only to authenticated HROS internal staff (no public registration)
2. Tenant data strictly isolated — no cross-tenant data access possible from portal
3. Actual financial transactions processed by Stripe; portal manages metadata only
4. Audit logs are append-only — no admin (including Super Admin) can modify or delete them
5. Tenant Code is immutable after creation — no override mechanism

### 5.2 Key Assumptions `[A]`
1. HROS Core Backend microservices exist and expose stable REST APIs
2. Email delivery infrastructure is provisioned and operational
3. Stripe webhook events are handled by Billing Service, not this portal directly
4. Browser clients have reliable internet connectivity (no offline mode)
5. Redis failure degrades session performance but does not break core functionality

---

*End of SRS | HROS-SRS-001 v1.0*