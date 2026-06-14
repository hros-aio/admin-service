# Product Requirements Document (PRD)
## HROS Admin — Super Admin Portal
**Version:** 1.0  
**Document Type:** Product Requirements Document  
**Prepared by:** Senior Product Manager & SaaS Solution Architect  
**Source:** Reverse-engineered from UI/UX design screens  
**Classification:** Internal — Confidential  

> **Assumption Legend:** Items marked `[ASSUMED]` are inferred from standard SaaS HRMS behavior where the design did not explicitly specify the requirement.

---

## Table of Contents

1. [Product Overview](#1-product-overview)
2. [Business Goals](#2-business-goals)
3. [User Roles](#3-user-roles)
4. [Navigation Structure](#4-navigation-structure)
5. [Functional Requirements](#5-functional-requirements)
   - 5.1 Authentication
   - 5.2 Dashboard
   - 5.3 Tenant Management
   - 5.4 Subscription Management
   - 5.5 Plan Management
   - 5.6 Admin Management
   - 5.7 Policy Management
6. [Validation Rules](#6-validation-rules)
7. [Business Rules](#7-business-rules)
8. [Status Definitions](#8-status-definitions)
9. [Non-Functional Requirements](#9-non-functional-requirements)
10. [Audit Requirements](#10-audit-requirements)
11. [Future Enhancements](#11-future-enhancements)

---

## 1. Product Overview

HROS Admin is a **Super Admin Portal** for the HROS (Human Resources Operating System) SaaS platform. It is a centralized, multi-tenant management console used exclusively by internal HROS operators — not by end-customer tenants. The portal allows HROS staff to provision and manage client organizations (tenants), configure subscription plans, control access via role-based policies, and oversee the health of the entire platform ecosystem.

The portal operates at the infrastructure layer above all tenant-facing products, giving authorized HROS administrators god-mode visibility and control across all customer accounts.

**Current System Version:** v4.8.2-stable (observed from dashboard footer)  
**Platform:** Web-based, browser-accessible  
**Primary Users:** HROS internal operations, billing, and support teams  

---

## 2. Business Goals

| # | Goal | Rationale |
|---|------|-----------|
| BG-01 | Reduce tenant onboarding time | Streamlined tenant creation with wizard-style form reduces manual provisioning effort |
| BG-02 | Centralize subscription lifecycle management | Full plan upgrade/downgrade/cancel/trial controls from one portal |
| BG-03 | Enforce zero-trust security architecture | Role-based policy engine with conflict detection ensures least-privilege access |
| BG-04 | Enable real-time operational visibility | Dashboard with live tenant activity feed and subscription trend analytics |
| BG-05 | Support scalable multi-tenant SaaS growth | Architecture supports 1,200+ tenants with structured plan tiers and limit management |
| BG-06 | Ensure regulatory compliance | GDPR-aware policy conditions, audit log retention, and export controls |
| BG-07 | Minimize revenue leakage | Automated subscription renewal, expiry tracking, and MRR visibility |

---

## 3. User Roles

### 3.1 Super Admin
- Full system-wide access to all modules
- Can create, update, archive tenants and plans
- Can invite and manage all admin accounts
- Can define and modify all system policies
- Can access all audit logs
- Can override tenant quotas and billing settings

### 3.2 Manager
- Access to standard operations and employee records
- Can view tenants and subscriptions
- Cannot modify system-level policies
- `[ASSUMED]` Cannot archive tenants or delete plans

### 3.3 Auditor
- Read-only access for reporting and compliance checks
- Can view all tenant and subscription data
- Can export audit logs and reports
- Cannot create, update, or delete any entities

### 3.4 Support Lead `[ASSUMED]`
- Can handle tickets and tenant detail queries
- Read/update on tenant records
- Cannot manage plans or policies

### 3.5 Billing Admin `[ASSUMED]`
- Can manage plans, invoices, and payments
- Cannot invite admins or modify system policies

### 3.6 Content Editor `[ASSUMED]`
- Can update policy documentation and help articles
- No access to tenant financial data

### 3.7 Audit Specialist `[ASSUMED]`
- Read-only access to logs and exports
- No create/update/delete permissions on any entity

---

## 4. Navigation Structure

```
HROS Admin Portal
├── Dashboard
├── Tenants
│   ├── Tenant List
│   ├── Create Tenant
│   ├── Update Tenant
│   ├── Archive Tenant [action]
│   └── Tenant Subscription Detail
│       └── Manage Subscription
├── Plans
│   ├── Plan List
│   ├── Create Plan
│   └── Edit Plan (slide-in panel)
├── Admins
│   └── Admin List + Invite Admin (slide-in panel)
├── Policies
│   └── Policy Management + Permission Matrix
├── Settings [ASSUMED]
├── Support
└── Logout
```

**Top Bar (persistent):**
- Global search bar (context-sensitive placeholder text per module)
- Notification bell with unread badge
- Help / question mark icon
- Logged-in admin avatar + name + role label

---

## 5. Functional Requirements

---

### 5.1 Authentication

**Business Purpose:** Secure access to the HROS Super Admin Portal, preventing unauthorized access to sensitive multi-tenant customer and billing data.

**Screen:** `hros_admin_login.png`

#### 5.1.1 Features

| Feature ID | Feature | Description |
|------------|---------|-------------|
| AUTH-01 | Email/Password Login | Standard credential-based login with masked password field |
| AUTH-02 | Password Visibility Toggle | Eye icon to reveal/hide password |
| AUTH-03 | Forgot Password | Link to initiate password reset flow |
| AUTH-04 | Persistent Session | "Keep me logged in for 30 days" checkbox |
| AUTH-05 | SSO Login | Single Sign-On via enterprise identity provider |
| AUTH-06 | Biometric Login | Fingerprint/device biometric authentication option |
| AUTH-07 | IT Support Contact | Link to Contact IT Support for system administrators |

#### 5.1.2 User Actions
- Enter email address
- Enter password
- Toggle password visibility
- Check "Keep me logged in" (optional)
- Click "Sign In to Dashboard"
- Click "Forgot Password?" to initiate reset
- Click "SSO" for enterprise SSO flow
- Click "Biometric" for device-based auth

#### 5.1.3 Validations
- Email must be valid format (RFC 5322)
- Email domain must match authorized HROS admin domain `[ASSUMED: @hros.com or @hros.admin]`
- Password must not be blank
- `[ASSUMED]` After 5 failed attempts, account is temporarily locked
- `[ASSUMED]` MFA required for Super Admin role logins

#### 5.1.4 Success Criteria
- Valid credentials redirect to Dashboard
- Session persisted for 30 days if checkbox selected
- Invalid credentials show inline error without revealing which field is wrong

---

### 5.2 Dashboard

**Business Purpose:** Provide a real-time operational overview of the HROS platform's health, tenant activity, and subscription performance.

**Screen:** `dashboard_overview.png`

#### 5.2.1 KPI Cards

| Metric | Value (Sample) | Indicator |
|--------|---------------|-----------|
| Total Tenants | 1,248 | +12% growth indicator |
| Active Tenants | 1,192 | "Stable" label |
| Expired Subscriptions | 24 | -4% trend |
| Total Admin Users | 5,842 | — |
| Active Plans | 14 | — |

#### 5.2.2 Features

| Feature ID | Feature | Description |
|------------|---------|-------------|
| DASH-01 | KPI Summary Cards | Five top-level metric cards with trend indicators |
| DASH-02 | Subscription Status Trends Chart | Line/bar chart showing subscription growth and retention over 6M or 1Y |
| DASH-03 | Quick Actions Panel | Shortcuts to Create Tenant, Create Plan, Invite Admin, Manage Policies |
| DASH-04 | Recent Tenant Activity Feed | Live feed of last N administrative actions with tenant, action, status, datetime, and operator |
| DASH-05 | Export Report | Export dashboard data as a report |
| DASH-06 | New Command | `[ASSUMED]` Command palette / shortcut launcher |
| DASH-07 | View All Activities | Navigate to full activity log |

#### 5.2.3 Recent Activity Feed Columns
- Tenant name + avatar initials
- Action description (e.g., "Subscription Renewed", "Policy Update: GDPR-24", "Seat Count Increased (+50)")
- Status badge (Success / Applied / Pending)
- Date/Time
- Operator (user or system process that triggered the action)

#### 5.2.4 Chart Toggle
- 6-Month view (default)
- 1-Year view

#### 5.2.5 User Actions
- Toggle chart between 6M / 1Y
- Click any Quick Action to navigate
- Click "View All Activities" to open full audit log
- Click "Export Report" to download dashboard summary

#### 5.2.6 Success Criteria
- All KPI cards load within 2 seconds
- Activity feed reflects actions within 60 seconds of occurrence `[ASSUMED]`
- Chart renders with correct data for selected period

---

### 5.3 Tenant Management

**Business Purpose:** Manage the full lifecycle of client organizations (tenants) on the HROS platform — including onboarding, configuration, updating, and archiving.

#### 5.3.1 Tenant List

**Screen:** `tenant_management.png`

##### Features

| Feature ID | Feature | Description |
|------------|---------|-------------|
| TM-01 | Tenant Table | Paginated list of all tenants with key attributes |
| TM-02 | Status Filter | Dropdown filter: All Statuses / Active / Suspended / Pending / Expired |
| TM-03 | Plan Type Filter | Dropdown filter by subscription plan type |
| TM-04 | Date Range Filter | Filter by tenant creation date range |
| TM-05 | Clear Filters | Reset all active filters |
| TM-06 | Advanced Filters | Additional filter options (column icon button) |
| TM-07 | Create Tenant Button | Navigate to Create Tenant form |
| TM-08 | Pagination | 10 results per page `[ASSUMED]`, with page navigation |
| TM-09 | Summary Stats Footer | Active Tenants, Total MRR, Suspended count, Expiring in 30d |

##### Table Columns
- Tenant Name (with avatar initials, sub-location label)
- Code (unique tenant code, e.g., GN-8842)
- Owner Email
- Plan (badge: Enterprise / Professional / Basic)
- Status (badge: Active / Suspended / Pending / Expired)
- Created date
- Expiry date

##### Footer Summary Cards
| Card | Sample Value |
|------|-------------|
| Active Tenants | 112 |
| Total MRR | $248.5k |
| Suspended | 9 |
| Expiring (30d) | 14 |

---

#### 5.3.2 Create Tenant

**Screen:** `create_tenant.png`

**Business Purpose:** Provision a new client organization on the HROS platform, setting up their identity, admin credentials, subscription, and operational defaults in a single workflow.

##### Form Sections

**Section 1: Tenant Information**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| Tenant Name | Text | Yes | e.g., "Acme Corp Global" |
| Tenant Code | Text | Yes | e.g., "ACME_001" — likely auto-generated with override `[ASSUMED]` |
| Legal Company Name | Text | No | Official registered business name |
| Industry | Dropdown | No | Select from predefined industry list |
| Country | Dropdown | No | Default: United States |
| Timezone | Dropdown | No | Default: (GMT-05:00) Eastern Time |
| Status | Radio/Card Select | No | Active / Pending / Suspended |

**Section 2: Owner Information**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| Owner Name | Text | Yes | Full legal name of account owner |
| Owner Email | Email | Yes | Primary billing/contact email |
| Owner Phone | Phone | No | `[ASSUMED]` |

**Section 3: Initial Admin Account**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| Admin Name | Text | Yes | Internal admin display name |
| Admin Email | Email | Yes | Login email for tenant's first admin |
| Admin Password | Password | Yes `[ASSUMED]` | Auto-generated or set manually |
| Force Password Reset on First Login | Checkbox | No | Default: unchecked |

**Section 4: Subscription Setup**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| Select Plan | Dropdown | Yes | e.g., "Professional Plus" |
| Billing Cycle | Toggle | No | Monthly (default) / Yearly |
| Start Date | Date picker | No | |
| End Date | Date picker | No | |
| Enable Trial Period | Checkbox | No | When checked, reveals Trial End Date |
| Trial End Date | Date picker | Conditional | Required if trial enabled |

**Section 5: Tenant Settings**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| Default Language | Dropdown | No | Default: English (US) |
| Default Currency | Dropdown | No | Default: USD ($) |
| Limit Overrides — Employees | Number | No | Override plan's employee cap |
| Limit Overrides — Admins | Number | No | Override plan's admin seat cap |
| Internal Deployment Notes | Textarea | No | Custom configs, SLA agreements, implementation notes |

##### User Actions
- Fill all required fields
- Select tenant status
- Toggle billing cycle
- Enable/disable trial period
- Click "Create Tenant" to submit
- Click "Cancel" to discard

##### Validations
- Tenant Name: required, max 100 chars, unique `[ASSUMED]`
- Tenant Code: required, alphanumeric + underscore, unique across system
- Owner Email: valid email format, unique `[ASSUMED]`
- Admin Email: valid email format, different from Owner Email `[ASSUMED]`
- If trial enabled, Trial End Date must be after Start Date
- End Date must be after Start Date
- All asterisked (*) fields mandatory for system instantiation (shown in footer note)

##### System Indicators
- "SYSTEM READY" badge in top-right confirms platform readiness before tenant creation

##### Success Criteria
- Tenant provisioned and visible in tenant list
- Initial admin receives onboarding email `[ASSUMED]`
- Subscription activated per configured plan and dates

---

#### 5.3.3 Update Tenant

**Screen:** `update_tenant.png`

**Business Purpose:** Modify an existing tenant's organizational configuration, admin credentials, subscription assignment, and operational settings.

##### Key Differences from Create Tenant
- Page title: "Update Tenant Details"
- Breadcrumb shows: Tenants > Update Tenant
- Fields pre-populated with existing values (e.g., Tenant Name = "Global Logistics Inc.", Code = "GL-8842", Industry = "Logistics")
- Two additional action buttons in header:
  - **Archive Tenant** (red/destructive) — initiates archival workflow
  - **System Ready** (status indicator badge)
- Submit button label: "Save Changes"
- Cancel button present

##### Additional Fields (vs. Create)
- Owner Phone: visible and pre-populated
- Admin Password field present (for credential reset)
- Force password reset on first login checkbox

##### Workflow
1. Admin navigates to tenant record
2. Portal pre-fills all fields with current data
3. Admin modifies desired fields
4. Admin clicks "Save Changes"
5. `[ASSUMED]` Changes trigger audit log entry and confirmation notification

##### Validations
- Same as Create Tenant
- `[ASSUMED]` Changing subscription plan mid-cycle triggers prorated billing warning
- `[ASSUMED]` Downgrading to lower plan validates current usage doesn't exceed new plan limits

---

#### 5.3.4 Archive Tenant

**Business Purpose:** Soft-delete a tenant organization, preserving data for compliance while removing active access.

##### Trigger
- "Archive Tenant" button on Update Tenant page

##### Expected Behavior `[ASSUMED]`
- Confirmation modal with warning message
- Requires reason for archival
- Sets tenant status to "Archived"
- Suspends all active subscriptions
- Prevents tenant admin logins
- Data retained for configurable period (e.g., 90 days) per data retention policy
- Irreversible without Super Admin intervention
- Creates audit log entry with operator, timestamp, and reason

##### Permissions
- Only Super Admin role can archive tenants

---

#### 5.3.5 Tenant Subscription Detail

**Screen:** `tenant_subscription_detail.png`

**Business Purpose:** View the complete subscription state for a specific tenant, including active plan, resource utilization, billing history, and payment method.

##### Breadcrumb
Tenants > Acme Corp > Subscription

##### Plan Summary Card
- Plan name: Enterprise Growth Plan
- Description text
- Status badge: ACTIVE (green)
- Billing Cycle: Annually
- Start Date: Jan 12, 2023
- Next Renewal: Jan 12, 2024
- Amount: $12,499.00/yr

##### Payment Method Panel
- Card brand + last 4 digits (e.g., VISA ••••8829)
- Card expiry
- Edit link
- View Billing Address button
- Manage Subscription button (navigates to Manage Subscription page)

##### Resource Usage Section
| Metric | Sample | Percentage |
|--------|--------|-----------|
| Employees | 854 / 1,000 | 85% |
| Admins | 12 / 20 | 60% |
| Cloud Storage | 46 GB / 50 GB | 92% (red — near limit) |
| Monthly API Calls | 1.2M / 5M | 24% |

- Updated timestamp shown ("Updated 5 mins ago")
- Color-coded progress bars (blue = normal, orange = warning, red = critical) `[ASSUMED thresholds: >90% = red, >70% = orange]`

##### Billing History Table

| Column | Notes |
|--------|-------|
| Invoice ID | Alphanumeric ID (e.g., INV-2023-001) |
| Date | Invoice date |
| Amount | Invoice total |
| Status | Paid (green) / Refunded (red) / Pending `[ASSUMED]` |
| Action | View/download invoice icon |

- "Download All Invoices" link

##### Footer Action Tiles
- Change Billing Cycle — Switch between monthly and annual
- Usage Reports — Export detailed consumption data
- Cancel Plan — End subscription at period close

##### Header Actions
- Pause Subscription button
- Upgrade Plan button

---

### 5.4 Subscription Management

#### 5.4.1 Manage Subscription

**Screen:** `manage_tenant_subscription.png`

**Business Purpose:** Adjust billing configurations, plan assignments, quota overrides, and trial management for an existing tenant subscription.

##### Warning Banner
"Sensitive Action Required — Changes to employee limits or billing cycles may trigger immediate prorated charges or restrict access for existing users. Please verify with the account manager before proceeding."

##### Current Subscription Summary
- Current Plan name (e.g., Enterprise Plus) displayed prominently
- Employees used / total (progress bar)
- Storage used / total (progress bar)
- Billing amount per month
- Status badge (ACTIVE)

##### Billing Settings Panel
| Setting | Control |
|---------|---------|
| Billing Cycle | Toggle: Monthly / Yearly |
| Auto-Renew | Checkbox: Automatic charge on renewal date |
| Payment Status | Verified / Unverified indicator |
| Last Invoice | Invoice number + payment status |

##### Change Subscription Plan
| Field | Type | Notes |
|-------|------|-------|
| Select New Plan | Dropdown | Shows all available plans including custom |
| Effective Date | Dropdown | "Immediately" or scheduled date `[ASSUMED]` |
| Downgrade Plan | Button (outline) | Triggers downgrade confirmation |
| Upgrade Plan | Button (filled) | Triggers upgrade confirmation |

##### Manual Quota Overrides
| Override | Field |
|----------|-------|
| Employee Limit | Numeric stepper |
| Admin Seats | Numeric stepper |
| Storage (GB) | Numeric stepper |
| API Calls / Month | Numeric stepper |

##### Trial Management Panel
- Trial Status: INACTIVE / ACTIVE badge
- End Date field
- Enable Trial button
- Extend Trial button

##### Internal Admin Notes
- Free-text area for private subscription notes
- "Visible only to super administrators" note

##### Header Actions
- Pause Subscription button
- Cancel Subscription button (destructive)

##### Unsaved Changes Notification
- Sticky bottom panel: "Unsaved Changes" with "Save All Settings" and "Discard Draft" buttons

#### 5.4.2 Upgrade Plan

**Business Purpose:** Move tenant to a higher-tier plan with expanded limits and features.

##### Workflow `[ASSUMED]`
1. Select new plan (higher tier) from dropdown
2. Select effective date (Immediately or next billing period)
3. Click "Upgrade Plan"
4. Confirmation modal with proration summary
5. Confirm → plan updated, new invoice generated, audit log entry created

##### Business Rules
- Upgrade takes effect immediately or at next billing cycle
- Prorated charges applied when upgrading mid-cycle
- Resource limits updated immediately upon confirmation

#### 5.4.3 Downgrade Plan

**Business Purpose:** Move tenant to a lower-tier plan with reduced limits.

##### Business Rules
- System must validate current usage against new plan limits before allowing downgrade
- If current usage exceeds new plan limits, downgrade is blocked with error message `[ASSUMED]`
- Downgrade typically takes effect at end of current billing period `[ASSUMED]`
- Credit may be applied as account balance `[ASSUMED]`

#### 5.4.4 Trial Management

**Business Purpose:** Enable, extend, or manage trial periods for tenant subscriptions.

| Action | Behavior |
|--------|----------|
| Enable Trial | Activates trial period; sets trial end date |
| Extend Trial | Pushes back trial end date; requires new end date input |
| Trial Status: INACTIVE | Trial not in use or already expired |
| Trial Status: ACTIVE | Trial currently running `[ASSUMED]` |

##### Business Rules
- Trial period cannot exceed plan's configured maximum trial days
- `[ASSUMED]` Trial can only be enabled once per tenant per plan
- `[ASSUMED]` System sends reminder notifications at 3 days and 1 day before trial end
- Extending trial requires Super Admin or Billing Admin role

#### 5.4.5 Cancel Subscription / Pause Subscription

**Pause Subscription:**
- Temporarily halts billing and access
- `[ASSUMED]` Maximum pause duration: 90 days
- `[ASSUMED]` Tenant data preserved; access suspended
- Requires confirmation

**Cancel Subscription:**
- Terminates subscription at end of current billing period
- `[ASSUMED]` Cancellation confirmation dialog with impact summary
- Triggers data retention countdown
- Creates audit log entry

---

### 5.5 Plan Management

**Business Purpose:** Define, configure, and maintain the subscription tier catalog that tenants subscribe to.

#### 5.5.1 Plan List

**Screen:** `plan_management.png`

##### Plan Cards Display
Each plan shown as a card with:
- Tier label (STARTER / SCALE / UNLIMITED)
- Plan name (Basic / Pro / Enterprise)
- Monthly price (billed annually)
- Max Employees
- Max Admins
- Storage
- Feature inclusions list (with checkmarks and crossed-out unavailable features)
- "Edit Plan Details" button
- "Most Popular" badge (on featured plan)

##### Sample Plans

| | Basic | Pro | Enterprise |
|--|-------|-----|-----------|
| Label | STARTER | SCALE | UNLIMITED |
| Price | $49/mo | $199/mo | $499/mo |
| Max Employees | 50 | 500 | Unlimited |
| Max Admins | 2 | 10 | Unlimited |
| Storage | 10 GB | 100 GB | 1 TB |
| Highlighted Feature | — | Most Popular | — |

##### Module Availability Status Section
Shows which modules are available per plan tier:
- Employee Mgmt: ALL PLANS
- Payroll: PRO + ENT
- Reports: PRO + ENT
- API Access: ENTERPRISE ONLY

##### Create New Plan Button
Navigates to Create Plan form.

---

#### 5.5.2 Create Plan

**Screen:** `create_plan.png`

**Business Purpose:** Define a new subscription tier with pricing, resource limits, feature permissions, and administrative settings.

##### Form Sections

**Section 1: Basic Information**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| Plan Name | Text | Yes | e.g., "Enterprise Plus" |
| Plan Code | Text | Yes | e.g., "EP_2024_01" — unique identifier |
| Description | Textarea | No | Target audience and value proposition |
| Plan Status | Toggle | No | Default ON — controls visibility for new signups |

**Section 2: Pricing & Billing**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| Currency | Dropdown | Yes | Default: USD ($) |
| Monthly Price | Number | Yes | Default: $0.00 |
| Yearly Price | Number | Yes | Default: $0.00 |
| Trial Period (Days) | Number | No | Default: 14 |

**Section 3: Capacity & Resource Limits**

| Field | Default | Unit |
|-------|---------|------|
| Max Locations | 3 | — |
| Storage Limit | 50 | GB |
| API Rate Limit | 1000 | req/min |
| Max Employees | — | `[ASSUMED]` |
| Max Admin Users | — | `[ASSUMED]` |
| Max Companies | — | `[ASSUMED]` |

**Section 4: Feature Permissions**

| Feature | Description | Configurable |
|---------|-------------|-------------|
| Employee Mgmt | Core records | Checkbox |
| Attendance | Clock-in/out | Checkbox |
| Leave Management | Time-off requests | Checkbox |
| Payroll | Salary & Taxes | Checkbox |
| Expense Mgmt | Claim processing | Checkbox |
| Asset Management | Hardware tracking | Checkbox |
| Reports | Basic analytics | Checkbox |
| API Access | Integration keys | Checkbox |

**Section 5: Advanced Administration**

| Field | Options | Notes |
|-------|---------|-------|
| Support Level | Basic / Premium / Enterprise | Toggle group |
| Audit Log Retention | Number (days) | Default: 365 |
| Allow Custom Branding | Checkbox | Tenants upload own logos/brand colors |
| Enable Bulk Data Export | Checkbox | Export employee and payroll records to CSV/PDF |

##### Footer Note
"Plan configuration will be active immediately after creation."

##### User Actions
- Complete form sections
- Toggle plan status
- Select feature permissions
- Click "Create Plan" to publish
- Click "Cancel" to discard

---

#### 5.5.3 Update Plan

**Screen:** `update_plan.png`

**Business Purpose:** Modify existing plan details, pricing, limits, or features while managing downstream impact on existing tenant subscriptions.

##### Key Differences from Create Plan
- Page title: "Edit Subscription Plan"
- Warning banner: "Note: Changes to pricing or capacity limits may affect existing tenant subscriptions. Please verify before saving."
- Fields pre-populated with existing plan data (e.g., Plan Name = "Pro Plan", Code = "PRO_01")
- Submit button: "Save Changes" instead of "Create Plan"

##### Inline Edit Panel (from Plan List)
The Plan Management screen also offers a slide-in right panel for quick edits:
- Plan Name field
- Monthly Price
- Yearly Discount (%)
- Max Employees / Max Admins
- Module toggles: Employee Management, Attendance & Leave, Payroll Module, API Access
- Save Changes / Cancel buttons

#### 5.5.4 Archive Plan `[ASSUMED]`

**Business Purpose:** Retire a plan so no new tenants can subscribe to it, while preserving existing subscriptions.

**Expected Behavior:**
- Plan status set to Archived/Inactive
- Existing tenants on the plan are unaffected
- Plan no longer appears in tenant signup or plan selection dropdowns
- Requires Super Admin confirmation
- Audit log entry created

---

### 5.6 Admin Management

**Business Purpose:** Manage HROS internal administrator accounts — including inviting new admins, assigning roles, and monitoring activity.

**Screen:** `admin_management.png`

#### 5.6.1 Admin List

##### KPI Cards
| Metric | Sample |
|--------|--------|
| Total Admins | 24 (+2 this month) |
| Super Admins | 4 (System-wide access) |
| Active Now | 12 (Normal system load) |

##### Admin Table Columns
| Column | Notes |
|--------|-------|
| Administrator | Name + email address |
| Role | Super Admin / Manager / Auditor |
| Status | Active (green) / Inactive (gray) badge |
| Last Login | Relative time (e.g., "2 mins ago", "5 hours ago") |
| Created Date | Full date |
| Actions | Three-dot kebab menu |

##### Actions Menu (per admin row) `[ASSUMED]`
- Edit admin details
- Change role
- Deactivate / reactivate account
- Reset password
- View activity log
- Remove admin

#### 5.6.2 Invite Administrator

##### Slide-in Panel Fields

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| Full Name | Text | Yes | e.g., "Sarah Johnson" |
| Work Email | Email | Yes | e.g., "sarah.j@hros.com" |
| Administrative Role | Radio/Card select | Yes | Super Admin / Manager / Auditor |
| Account Status | Toggle | No | Default: Active |

##### Role Descriptions (as shown in UI)
- **Super Admin:** Full access to all system settings and user management
- **Manager:** Access to standard operations and employee records
- **Auditor:** Read-only access for reporting and compliance checks

##### User Actions
- Fill required fields
- Select role from card-style selector
- Toggle initial account status
- Click "Send Invitation" — sends email invite
- Click "Cancel" to close panel

##### Validations
- Work Email: valid format, must not duplicate existing admin email
- Full Name: required, max 100 chars `[ASSUMED]`
- Role: required selection before submission

##### Business Rules
- Only Super Admin can invite new Super Admins `[ASSUMED]`
- Invitation email expires after 48 hours `[ASSUMED]`
- New admin must set password on first login `[ASSUMED]`
- Account Status toggle sets initial state; can be changed later

##### Success Criteria
- Invited admin appears in table with "Pending" or "Inactive" status until they accept
- Invitation email delivered to work email address
- Audit log records invitation with inviter's identity

---

### 5.7 Policy Management

**Business Purpose:** Define and enforce role-based access control (RBAC) across the HROS platform, configure permission matrices, build policy condition logic, and monitor security posture.

**Screen:** `policy_management.png`

#### 5.7.1 System Roles Panel (Left)

Displays all system-defined roles:
- Super Admin — Full system access & control
- Support Lead — Handle tickets & tenant details
- Billing Admin — Manage plans, invoices & payments
- Content Editor — Policy updates & help articles
- Audit Specialist — Read-only logs & exports
- "+ Add New Role" button (dashed outline)

Selecting a role loads its permission matrix on the right.

#### 5.7.2 Permission Matrix

Displayed as a grid table:

##### Columns (Permission Types)
View | Create | Update | Delete | Approve | Export

##### Rows (Modules)
- Tenant Management
- Plan Management
- Admin Management
- Policy Management

Each cell is a checkbox. Global select buttons: "All" / "None" at table header.

**Sample: Super Admin** — All permissions checked across all modules.

#### 5.7.3 Policy Condition Builder

Allows building custom access logic conditions:

| Field | Control | Example |
|-------|---------|---------|
| Subject | Dropdown | Super Admin |
| Action | Dropdown | Create Tenant |
| Condition | Text input | e.g., "location == 'US'" |
| Effect | Radio | Allow / Deny |

##### Active Logic Chains (displayed as cards)
- "If Subject is *Super Admin*, **Allow** all Actions." — [System Default]
- "If Location is *Outside EU*, **Deny** Export CSV." — [GDPR Compliance]

##### User Actions
- Select role from left panel
- Modify permission checkboxes in matrix
- Build new policy conditions using builder fields
- Click "+ Append New Logic Statement" to add condition
- Click "Save Changes" to apply
- Click "Audit Logs" to view policy change history

#### 5.7.4 Real-time Policy Conflict Detection

- System automatically analyzes overlapping permissions to prevent "Privilege Creep"
- Ensures zero-trust security architecture across all tenants
- "Run System Health Check" link triggers manual validation

##### Security Score Panel
- Current Score: 98.2 (+1.4%)
- Description: "Policies currently cover 100% of critical endpoints. Excellent standing."

#### 5.7.5 Business Rules
- Policies are evaluated at runtime for every API call `[ASSUMED]`
- More specific policies override general ones `[ASSUMED]`
- GDPR-tagged policies cannot be modified without Super Admin approval `[ASSUMED]`
- Policy changes are versioned and audited
- Conflict detection runs after every save

---

## 6. Validation Rules

| Rule ID | Module | Field | Rule | Error Message |
|---------|--------|-------|------|---------------|
| VAL-01 | All | Required fields | Must not be empty | "This field is required" |
| VAL-02 | Auth | Email | RFC 5322 format | "Enter a valid email address" |
| VAL-03 | Create Tenant | Tenant Code | Alphanumeric + underscore, unique | "Tenant code already exists" |
| VAL-04 | Create Tenant | Admin Email | Valid format, unique across admins | "Email already in use" |
| VAL-05 | Create Tenant | Trial End Date | Must be after Start Date | "Trial end date must be after start date" |
| VAL-06 | Create Tenant | End Date | Must be after Start Date | "End date must be after start date" |
| VAL-07 | Create Plan | Plan Code | Alphanumeric + underscore, unique | "Plan code already exists" |
| VAL-08 | Create Plan | Monthly Price | Numeric, ≥ 0 | "Enter a valid price" |
| VAL-09 | Create Plan | Trial Period | Integer, ≥ 0, ≤ 365 | "Trial period must be between 0 and 365 days" |
| VAL-10 | Invite Admin | Work Email | Valid format, not duplicate | "An admin with this email already exists" |
| VAL-11 | Subscription | Downgrade | Usage ≤ new plan limits | "Current usage exceeds limits of selected plan" |
| VAL-12 | Policy | Condition | Valid logical expression syntax | "Invalid condition syntax" |
| VAL-13 | Create Tenant | Owner Email | Valid email format | "Enter a valid email address" |
| VAL-14 | Manage Subscription | Quota Overrides | Numeric, ≥ plan default | "Override cannot be less than plan minimum" `[ASSUMED]` |

---

## 7. Business Rules

| Rule ID | Rule | Applies To |
|---------|------|-----------|
| BR-01 | Tenant Code is globally unique and immutable after creation | Tenants |
| BR-02 | Only Super Admin can archive tenants | Tenants |
| BR-03 | Archived tenants cannot be onboarded with the same Tenant Code | Tenants |
| BR-04 | Subscription plan changes mid-cycle trigger prorated billing | Subscriptions |
| BR-05 | Downgrade is blocked if current usage exceeds new plan limits | Subscriptions |
| BR-06 | Trial can only be enabled once per tenant per plan | Subscriptions |
| BR-07 | Trial extension requires explicit admin action; no auto-extension | Subscriptions |
| BR-08 | Plan archival does not affect existing tenant subscriptions | Plans |
| BR-09 | Archived plans are invisible to new tenant signups | Plans |
| BR-10 | Plan Code is immutable after creation | Plans |
| BR-11 | Only Super Admin can invite other Super Admins | Admins |
| BR-12 | Admin invitation link expires after 48 hours | Admins |
| BR-13 | An admin's own account cannot be deactivated by themselves | Admins |
| BR-14 | GDPR policy conditions override all other allow rules for EU data | Policies |
| BR-15 | Policy changes require Save Changes action to take effect (not auto-saved) | Policies |
| BR-16 | Manual quota overrides supersede plan defaults for that tenant | Subscriptions |
| BR-17 | Resource usage at >90% triggers warning indicators `[ASSUMED]` | Subscriptions |
| BR-18 | Expired subscriptions restrict tenant access but preserve data | Subscriptions |
| BR-19 | Auto-Renew enabled subscriptions are charged automatically on renewal date | Subscriptions |
| BR-20 | Billing cycle changes take effect at start of next billing period `[ASSUMED]` | Subscriptions |
| BR-21 | Super Admin always has access to all modules regardless of policy matrix | Policies |
| BR-22 | Unsaved changes in subscription management are preserved as a draft until explicit save or discard | Subscriptions |

---

## 8. Status Definitions

### 8.1 Tenant Statuses

| Status | Description | Trigger |
|--------|-------------|---------|
| Active | Tenant is operational, subscription is current | Created with Active status or subscription renewed |
| Pending | Tenant setup in progress, awaiting activation | Created with Pending status or subscription not yet started |
| Suspended | Tenant access temporarily halted | Admin manual action, payment failure `[ASSUMED]` |
| Expired | Subscription period ended without renewal | End date passed without renewal |
| Archived | Tenant soft-deleted, data retained | Admin archive action |

### 8.2 Subscription Statuses

| Status | Description |
|--------|-------------|
| Active | Subscription is current and running |
| Trial | Within trial period |
| Inactive | Trial expired, subscription not converted `[ASSUMED]` |
| Paused | Temporarily suspended by admin |
| Cancelled | Marked for cancellation at period end |
| Expired | Subscription period ended |

### 8.3 Admin User Statuses

| Status | Description |
|--------|-------------|
| Active | Account enabled, can log in |
| Inactive | Account disabled, cannot log in |
| Pending `[ASSUMED]` | Invitation sent, not yet accepted |

### 8.4 Invoice/Payment Statuses

| Status | Description |
|--------|-------------|
| Paid | Payment successfully collected |
| Pending `[ASSUMED]` | Payment processing or awaiting |
| Refunded | Payment returned to customer |
| Failed `[ASSUMED]` | Payment attempt was unsuccessful |

### 8.5 Tenant Activity Statuses

| Status | Description |
|--------|-------------|
| Success | Action completed successfully |
| Applied | Policy or configuration successfully applied |
| Pending | Action initiated, awaiting completion |
| Failed `[ASSUMED]` | Action did not complete successfully |

---

## 9. Non-Functional Requirements

### 9.1 Performance

| Requirement | Target |
|-------------|--------|
| Dashboard load time | < 2 seconds for all KPI cards |
| Tenant list page load | < 1.5 seconds for first 10 results |
| Form submission response | < 3 seconds |
| Activity feed refresh rate | ≤ 60 seconds `[ASSUMED]` |
| Resource usage metrics freshness | ≤ 5 minutes (shown in UI) |
| System uptime | 99.9% SLA `[ASSUMED]` |

### 9.2 Security

| Requirement | Detail |
|-------------|--------|
| Authentication | Email/password + SSO + Biometric support |
| Session management | Configurable: session persists up to 30 days |
| MFA | `[ASSUMED]` Required for Super Admin accounts |
| Brute-force protection | `[ASSUMED]` Account lockout after 5 failed login attempts |
| Role-based access control | Enforced at UI and API level |
| Zero-trust architecture | Policy engine validates every action |
| Data encryption at rest | `[ASSUMED]` AES-256 |
| Data encryption in transit | TLS 1.2+ |
| GDPR compliance | Location-based export restrictions enforced via policy |
| Security score monitoring | Real-time scoring visible in Policy module |

### 9.3 Scalability

| Requirement | Target |
|-------------|--------|
| Tenant capacity | 1,200+ concurrent tenants (1,248 observed in production) |
| Admin users | 5,800+ total admin users supported |
| Active plans | 14+ concurrent plan configurations |
| Pagination | Tenant list supports 128+ pages |

### 9.4 Usability

| Requirement | Detail |
|-------------|--------|
| Responsive design | Web-based, browser-accessible `[ASSUMED]` |
| Form validation | Inline, real-time validation feedback |
| Confirmation dialogs | All destructive actions require confirmation |
| Warning banners | Sensitive actions display contextual warnings |
| Unsaved changes protection | Draft preservation with explicit save/discard |
| Breadcrumb navigation | Consistent across all sub-pages |

### 9.5 Compatibility

| Requirement | Detail |
|-------------|--------|
| Browser support | `[ASSUMED]` Modern browsers: Chrome, Firefox, Safari, Edge (latest 2 versions) |
| Minimum screen resolution | `[ASSUMED]` 1280×800 |

---

## 10. Audit Requirements

### 10.1 Events That Must Be Logged

| Event Category | Specific Events |
|----------------|-----------------|
| Authentication | Login success, login failure, logout, password reset, SSO login |
| Tenant Management | Create, update, archive, status change, quota override |
| Subscription Management | Plan change (upgrade/downgrade), pause, cancel, trial enable/extend, billing cycle change, payment event |
| Plan Management | Create, update, archive, status toggle |
| Admin Management | Invite, role change, activate, deactivate, remove |
| Policy Management | Permission matrix change, policy condition add/modify/delete, system health check run |
| Data Export | Any CSV/PDF/report export, "Download All Invoices" |

### 10.2 Audit Log Fields

Each audit log entry must capture:
- Event type
- Entity type (Tenant / Plan / Admin / Subscription / Policy)
- Entity ID and name
- Previous state `[ASSUMED]`
- New state `[ASSUMED]`
- Operator (user ID + name)
- Timestamp (UTC)
- IP address `[ASSUMED]`
- Source system (Auto-System, Billing Bot, or named admin user)

### 10.3 Audit Log Retention

- Default retention: 365 days (configurable per plan in Advanced Administration)
- `[ASSUMED]` Archived after retention period to cold storage, not deleted
- Accessible via "Audit Logs" button in Policy Management and from Dashboard Activity Feed

### 10.4 Audit Log Access

| Role | Access Level |
|------|-------------|
| Super Admin | Full read access to all audit logs |
| Auditor | Full read access to all audit logs |
| Audit Specialist | Full read access, export capability |
| Manager | `[ASSUMED]` Limited to their own actions |
| Billing Admin | `[ASSUMED]` Billing and subscription events only |

---

## 11. Future Enhancements

| Priority | Enhancement | Rationale |
|----------|-------------|-----------|
| High | Bulk Tenant Operations | Import/export tenant list via CSV for large-scale migrations |
| High | API Webhook Management | Allow admins to configure and test webhook endpoints per tenant |
| High | Automated Expiry Notifications | Email/SMS alerts for tenants and admins when subscriptions near expiry |
| Medium | Custom Plan Builder | Allow creation of fully custom plans with arbitrary feature combinations |
| Medium | Tenant Health Scoring | Automated health score per tenant based on usage, payment history, support tickets |
| Medium | Multi-currency Billing | Support billing in local currencies beyond USD |
| Medium | Advanced Analytics Dashboard | Cohort analysis, churn prediction, MRR trends with drill-down |
| Medium | Two-Factor Authentication (2FA) | TOTP/app-based 2FA as additional auth option beyond biometric |
| Low | White-label Portal Theming | Allow HROS staff to preview tenant white-label branding from admin portal |
| Low | In-portal Support Ticketing | Create/manage support tickets linked to tenant records directly |
| Low | AI-powered Anomaly Detection | Flag unusual usage spikes or suspicious admin activity automatically |
| Low | Tenant Communication Module | Send broadcast or targeted notifications to tenant admins from portal |
| Low | Mobile App for Super Admins | Native iOS/Android app for monitoring and lightweight admin tasks |

---

*End of Product Requirements Document*

---

**Document Version History**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2024 | Reverse-engineered from UI/UX designs | Initial draft |