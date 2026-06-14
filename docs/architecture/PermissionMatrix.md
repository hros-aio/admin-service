# Permission Matrix
## HROS Admin — Super Admin Portal
**Document ID:** HROS-PM-001 | **Version:** 1.0
**Classification:** Internal — Confidential
> `[A]` = Assumption inferred from SaaS HRMS best practice
> Legend: ✅ = Full Access | 👁 = View Only | ✏️ = View + Update | ⛔ = No Access | 🔒 = System-Enforced (cannot be changed)

---

## 1. Role Definitions

| Role | Code | Type | Description |
|------|------|------|-------------|
| Super Admin | `super_admin` | System | Full unrestricted access to all modules. Matrix is read-only — cannot be reduced. |
| Manager | `manager` | System | Standard operational access. Day-to-day tenant and subscription viewing. |
| Billing Admin | `billing_admin` | System | Plan and subscription financial management. No policy or admin access. |
| Support Lead | `support_lead` | System | Tenant record lookup and limited update for customer support tasks. |
| Auditor | `auditor` | System | Read-only access across all modules + audit log export. |
| Content Editor | `content_editor` | System | Policy documentation only. Minimal data access. |
| Audit Specialist | `audit_specialist` | System | Audit log read + export only. No operational access. |
| Custom Role | `custom_*` | Custom | Created by Super Admin via Policy Management. Starts with zero permissions. |

---

## 2. Module Permission Matrix

### 2.1 Dashboard Module

| Permission | Super Admin | Manager | Billing Admin | Support Lead | Auditor | Content Editor | Audit Specialist |
|------------|:-----------:|:-------:|:-------------:|:------------:|:-------:|:--------------:|:----------------:|
| **View** (KPI cards, activity feed, trend chart) | ✅🔒 | ✅ | ✅ | ✅ | ✅ | ⛔ | ⛔ |
| **Create** (n/a) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Update** (n/a) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Delete** (n/a) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Approve** (n/a) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Export** (PDF/CSV report) | ✅🔒 | ⛔ | ⛔ | ⛔ | ✅ | ⛔ | ⛔ |

---

### 2.2 Tenant Management Module

| Permission | Super Admin | Manager | Billing Admin | Support Lead | Auditor | Content Editor | Audit Specialist |
|------------|:-----------:|:-------:|:-------------:|:------------:|:-------:|:--------------:|:----------------:|
| **View** (list, detail, search) | ✅🔒 | ✅ | ✅ `[A]` | ✅ | ✅ | ⛔ | ⛔ |
| **Create** (provision new tenant) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Update** (edit tenant details) | ✅🔒 | ✅ `[A]` | ⛔ | ✅ | ⛔ | ⛔ | ⛔ |
| **Delete** (archive tenant) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Approve** (restore archived) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Export** (tenant list CSV) | ✅🔒 | ⛔ | ✅ `[A]` | ⛔ | ✅ | ⛔ | ⛔ |

**Notes:**
- Archive Tenant requires `tenant_management.can_delete` AND role = Super Admin (double-gated)
- Tenant Code is immutable for all roles once created
- Manager can update tenant details but cannot create or archive tenants

---

### 2.3 Subscription Management Module

| Permission | Super Admin | Manager | Billing Admin | Support Lead | Auditor | Content Editor | Audit Specialist |
|------------|:-----------:|:-------:|:-------------:|:------------:|:-------:|:--------------:|:----------------:|
| **View** (subscription detail, billing history, usage) | ✅🔒 | ⛔ | ✅ | ✅ | ✅ | ⛔ | ⛔ |
| **Create** (enable trial, create subscription on tenant create) | ✅🔒 | ⛔ | ✅ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Update** (change plan, billing cycle, auto-renew, quotas, notes) | ✅🔒 | ⛔ | ✅ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Delete** (cancel subscription) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Approve** (pause subscription) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Export** (download all invoices) | ✅🔒 | ⛔ | ✅ | ⛔ | ✅ | ⛔ | ⛔ |

**Notes:**
- Internal Admin Notes are visible only to Super Admin role regardless of view permission
- Quota Override (manual limits above plan defaults) requires Super Admin
- Cancel Subscription requires Super Admin role (mapped to `can_delete` with SA role check)
- Pause Subscription requires Super Admin role (mapped to `can_approve`)
- Billing Admin can upgrade/downgrade plans and manage billing cycle/auto-renew

---

### 2.4 Plan Management Module

| Permission | Super Admin | Manager | Billing Admin | Support Lead | Auditor | Content Editor | Audit Specialist |
|------------|:-----------:|:-------:|:-------------:|:------------:|:-------:|:--------------:|:----------------:|
| **View** (catalog, plan details, module matrix) | ✅🔒 | ⛔ | ✅ | ⛔ | ✅ | ⛔ | ⛔ |
| **Create** (create new plan) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Update** (edit plan, toggle status) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Delete** (archive plan) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Approve** (n/a) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Export** (plan catalog export) `[A]` | ✅🔒 | ⛔ | ✅ | ⛔ | ✅ | ⛔ | ⛔ |

**Notes:**
- Only Super Admin can Create, Update, or Archive plans
- Plan Code is immutable for all roles once created
- Cannot archive the last active plan (system-enforced constraint, not role-based)

---

### 2.5 Admin Management Module

| Permission | Super Admin | Manager | Billing Admin | Support Lead | Auditor | Content Editor | Audit Specialist |
|------------|:-----------:|:-------:|:-------------:|:------------:|:-------:|:--------------:|:----------------:|
| **View** (admin list, admin detail) | ✅🔒 | ⛔ | ⛔ | ⛔ | ✅ `[A]` | ⛔ | ⛔ |
| **Create** (invite new admin) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Update** (edit details, change role, activate/deactivate, reset password) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Delete** (remove admin) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Approve** (n/a) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Export** (admin list export) `[A]` | ✅🔒 | ⛔ | ⛔ | ⛔ | ✅ | ⛔ | ⛔ |

**Notes:**
- Admin Management is exclusively a Super Admin module for all write actions
- Self-protection rule: Admin cannot deactivate or remove their own account (enforced regardless of role)
- Inviting Super Admin role requires actor to also be Super Admin (secondary gate on `can_create`)

---

### 2.6 Policy Management Module

| Permission | Super Admin | Manager | Billing Admin | Support Lead | Auditor | Content Editor | Audit Specialist |
|------------|:-----------:|:-------:|:-------------:|:------------:|:-------:|:--------------:|:----------------:|
| **View** (permission matrix, conditions, security score) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ✅ `[A]` | ⛔ |
| **Create** (add role, add condition) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Update** (edit permissions, edit conditions) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Delete** (remove condition, remove custom role) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Approve** (run system health check) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Export** (n/a) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |

**Notes:**
- Super Admin permission matrix is always full-access and read-only (cannot be reduced by anyone)
- Policy Management is exclusively a Super Admin domain for all write actions
- Content Editor may view policy documentation as a display convention `[A]`

---

### 2.7 Audit Logs Module

| Permission | Super Admin | Manager | Billing Admin | Support Lead | Auditor | Content Editor | Audit Specialist |
|------------|:-----------:|:-------:|:-------------:|:------------:|:-------:|:--------------:|:----------------:|
| **View** (browse, filter audit log) | ✅🔒 | ⛔ | ⛔ | ⛔ | ✅ | ⛔ | ✅ |
| **Create** (system only — no manual creation) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Update** (immutable — blocked at DB level) | ⛔🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Delete** (immutable — blocked at DB level) | ⛔🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Approve** (n/a) | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| **Export** (CSV/PDF export of filtered log) | ✅🔒 | ⛔ | ⛔ | ⛔ | ✅ | ⛔ | ✅ |

**Notes:**
- Update and Delete are permanently blocked at database level via DB RULE — no role can bypass this
- Audit log entries are automatically created by the system for every mutating operation
- Export action itself is logged in audit_logs

---

## 3. Consolidated Role Summary Table

> ✅ = All permissions in module | 👁 = View only | ✏️ = View + Update | 📋 = View + Export | ⛔ = No Access

| Module | Super Admin | Manager | Billing Admin | Support Lead | Auditor | Content Editor | Audit Specialist |
|--------|:-----------:|:-------:|:-------------:|:------------:|:-------:|:--------------:|:----------------:|
| Dashboard | ✅🔒 | 👁 | 👁 | 👁 | 📋 | ⛔ | ⛔ |
| Tenant Management | ✅🔒 | ✏️ | 👁 | ✏️ | 📋 | ⛔ | ⛔ |
| Subscription Management | ✅🔒 | ⛔ | ✏️ | 👁 | 📋 | ⛔ | ⛔ |
| Plan Management | ✅🔒 | ⛔ | 👁 | ⛔ | 📋 | ⛔ | ⛔ |
| Admin Management | ✅🔒 | ⛔ | ⛔ | ⛔ | 👁 | ⛔ | ⛔ |
| Policy Management | ✅🔒 | ⛔ | ⛔ | ⛔ | ⛔ | 👁 | ⛔ |
| Audit Logs | ✅🔒 | ⛔ | ⛔ | ⛔ | 📋 | ⛔ | 📋 |

---

## 4. Sensitive Action Permission Gates

Some actions require elevated permission beyond the module-level matrix. These are double-gated:

| Action | Required Module Permission | Additional Requirement |
|--------|---------------------------|------------------------|
| Archive Tenant | `tenant_management.can_delete` | Actor role = Super Admin |
| Cancel Subscription | `subscription_management.can_delete` | Actor role = Super Admin |
| Pause Subscription | `subscription_management.can_approve` | Actor role = Super Admin |
| Override Tenant Quotas | `subscription_management.can_update` | Actor role = Super Admin |
| Add Internal Subscription Notes | `subscription_management.can_update` | Actor role = Super Admin (view also SA-only) |
| Invite Super Admin | `admin_management.can_create` | Actor role = Super Admin |
| Change Role to/from Super Admin | `admin_management.can_update` | Actor role = Super Admin |
| Remove Admin Account | `admin_management.can_delete` | Actor ≠ target (self-protection) |
| Deactivate Admin | `admin_management.can_update` | Actor ≠ target (self-protection) |
| Modify Policy Matrix | `policy_management.can_update` | Target role ≠ Super Admin (immutable) |
| Run System Health Check | `policy_management.can_approve` | Actor role = Super Admin |
| Restore Archived Tenant | `tenant_management.can_approve` | Actor role = Super Admin |

---

## 5. Policy Condition Evaluation Order

When a request hits the API, the RBAC engine evaluates in this exact order:

```
1. Check admin account status (active? not locked?)
2. Check JWT validity and expiry
3. Check GDPR-tagged policy conditions (highest priority)
4. Check DENY conditions for (subject, action) pair
5. Check ALLOW conditions for (subject, action) pair
6. Check module-level role_permissions (can_view / can_create / etc.)
7. Check sensitive action double-gates (role = Super Admin, self-protection)
8. Default: DENY (zero-trust — no match = no access)
```

**DENY always overrides ALLOW for the same subject+action combination.**

---

## 6. Security Score Impact by Role Configuration

Based on the security score formula (SRS-POL-003 / FS-7.2):

| Misconfiguration | Deduction |
|-----------------|-----------|
| Module with no permission defined for any role | -2 per module |
| Detected privilege conflict (overlapping ALLOW+DENY) | -5 per conflict |
| Super Admin without MFA enabled | -10 |
| No GDPR policy covering EU tenant data exports | -15 |
| Policy condition not reviewed in 90+ days `[A]` | -3 per stale policy |

**Target minimum score: 85** | **Excellent standing: ≥ 95**

---

## 7. Default Seed Permissions

The following permissions are seeded on system initialization. These represent the default configuration before any admin customization in the Policy Management module.

### Super Admin (all ✅, read-only in UI)
All modules: `can_view=T, can_create=T, can_update=T, can_delete=T, can_approve=T, can_export=T`

### Manager
| Module | View | Create | Update | Delete | Approve | Export |
|--------|------|--------|--------|--------|---------|--------|
| Dashboard | ✅ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Tenant Management | ✅ | ⛔ | ✅ | ⛔ | ⛔ | ⛔ |
| Subscription Management | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Plan Management | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Admin Management | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Policy Management | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Audit Logs | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |

### Billing Admin
| Module | View | Create | Update | Delete | Approve | Export |
|--------|------|--------|--------|--------|---------|--------|
| Dashboard | ✅ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Tenant Management | ✅ | ⛔ | ⛔ | ⛔ | ⛔ | ✅ |
| Subscription Management | ✅ | ✅ | ✅ | ⛔ | ⛔ | ✅ |
| Plan Management | ✅ | ⛔ | ⛔ | ⛔ | ⛔ | ✅ |
| Admin Management | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Policy Management | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Audit Logs | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |

### Support Lead
| Module | View | Create | Update | Delete | Approve | Export |
|--------|------|--------|--------|--------|---------|--------|
| Dashboard | ✅ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Tenant Management | ✅ | ⛔ | ✅ | ⛔ | ⛔ | ⛔ |
| Subscription Management | ✅ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Plan Management | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Admin Management | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Policy Management | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Audit Logs | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |

### Auditor
| Module | View | Create | Update | Delete | Approve | Export |
|--------|------|--------|--------|--------|---------|--------|
| Dashboard | ✅ | ⛔ | ⛔ | ⛔ | ⛔ | ✅ |
| Tenant Management | ✅ | ⛔ | ⛔ | ⛔ | ⛔ | ✅ |
| Subscription Management | ✅ | ⛔ | ⛔ | ⛔ | ⛔ | ✅ |
| Plan Management | ✅ | ⛔ | ⛔ | ⛔ | ⛔ | ✅ |
| Admin Management | ✅ | ⛔ | ⛔ | ⛔ | ⛔ | ✅ |
| Policy Management | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Audit Logs | ✅ | ⛔ | ⛔ | ⛔ | ⛔ | ✅ |

### Content Editor
| Module | View | Create | Update | Delete | Approve | Export |
|--------|------|--------|--------|--------|---------|--------|
| Dashboard | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Tenant Management | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Subscription Management | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Plan Management | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Admin Management | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Policy Management | ✅ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Audit Logs | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |

### Audit Specialist
| Module | View | Create | Update | Delete | Approve | Export |
|--------|------|--------|--------|--------|---------|--------|
| Dashboard | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Tenant Management | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Subscription Management | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Plan Management | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Admin Management | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Policy Management | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ | ⛔ |
| Audit Logs | ✅ | ⛔ | ⛔ | ⛔ | ⛔ | ✅ |

---

## 8. Cross-Reference: User Stories ↔ Permissions

| User Story | Required Permission(s) |
|------------|----------------------|
| US-DASH-01 | `dashboard.can_view` |
| US-DASH-05 | `dashboard.can_export` |
| US-TM-01, US-TM-02 | `tenant_management.can_view` |
| US-TM-03 | `tenant_management.can_view` + `subscription_management.can_view` |
| US-TM-04, US-TM-05 | `tenant_management.can_create` |
| US-TM-06 | `tenant_management.can_update` |
| US-TM-07 | `tenant_management.can_delete` AND Super Admin role |
| US-TM-08 | `tenant_management.can_view` |
| US-SUB-01 | `subscription_management.can_view` |
| US-SUB-02, US-SUB-03 | `subscription_management.can_update` |
| US-SUB-04 | `subscription_management.can_update` AND Super Admin role |
| US-SUB-05, US-SUB-06 | `subscription_management.can_create` or `can_update` |
| US-SUB-07 | `subscription_management.can_approve` AND Super Admin role |
| US-SUB-08 | `subscription_management.can_delete` AND Super Admin role |
| US-SUB-09, US-SUB-10 | `subscription_management.can_update` |
| US-SUB-11 | `subscription_management.can_update` AND Super Admin role |
| US-SUB-12 | `subscription_management.can_export` |
| US-PM-01 | `plan_management.can_view` |
| US-PM-02 | `plan_management.can_create` |
| US-PM-03, US-PM-04 | `plan_management.can_update` |
| US-PM-05 | `plan_management.can_delete` |
| US-PM-06 | `plan_management.can_update` |
| US-AM-01 | `admin_management.can_view` |
| US-AM-02 | `admin_management.can_create` |
| US-AM-03 | `admin_management.can_update` AND Super Admin role |
| US-AM-04 | `admin_management.can_update` AND actor ≠ target |
| US-AM-05 | `admin_management.can_update` |
| US-POL-01 | `policy_management.can_view` |
| US-POL-02 | `policy_management.can_update` |
| US-POL-03 | `policy_management.can_create` + `can_update` |
| US-POL-04 | `policy_management.can_view` |
| US-POL-05 | `policy_management.can_create` |
| US-POL-06 | `policy_management.can_approve` |
| US-AUD-01 | `audit_logs.can_view` |
| US-AUD-02 | `audit_logs.can_export` |
| US-AUD-03 | `audit_logs.can_view` + `policy_management.can_view` |

---

*End of Permission Matrix | HROS-PM-001 v1.0*