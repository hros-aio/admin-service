# User Stories
## HROS Admin — Super Admin Portal
**Document ID:** HROS-US-001 | **Version:** 1.0
**Format:** As a [role], I want to [action] so that [business value]
**Points:** Fibonacci (1, 2, 3, 5, 8, 13) | **Priority:** MH = Must Have, SH = Should Have, CH = Could Have

---

## Epic 1: Authentication & Session Management

### US-AUTH-01 — Credential Login
**As an** HROS admin user,
**I want to** log in using my work email and password,
**So that** I can securely access the Super Admin Portal and manage the platform.

**Role:** All admin roles | **Priority:** MH | **Points:** 3

**Notes:** Password field masked by default with toggle to reveal. Generic error on failure (no field specificity). Session token issued on success.

---

### US-AUTH-02 — SSO Login
**As an** HROS admin user,
**I want to** authenticate via my company's SSO provider,
**So that** I don't maintain a separate admin portal password and access is controlled centrally.

**Role:** All admin roles | **Priority:** MH | **Points:** 5

---

### US-AUTH-03 — Biometric Login
**As an** HROS admin user,
**I want to** log in using my device's biometric (fingerprint/face),
**So that** I can access the portal quickly on trusted devices without typing credentials.

**Role:** All admin roles | **Priority:** SH | **Points:** 3

---

### US-AUTH-04 — Forgot Password
**As an** HROS admin user,
**I want to** reset my password via a secure email link,
**So that** I can regain portal access without calling IT support.

**Role:** All admin roles | **Priority:** MH | **Points:** 3

---

### US-AUTH-05 — Persistent Session
**As an** HROS admin user,
**I want to** choose to stay logged in for 30 days on my work device,
**So that** I don't have to re-authenticate every work session.

**Role:** All admin roles | **Priority:** SH | **Points:** 2

---

### US-AUTH-06 — Account Lockout Alert
**As an** HROS admin user,
**I want to** receive an email alert when my account is locked due to failed login attempts,
**So that** I'm immediately aware of potential unauthorized access.

**Role:** All admin roles | **Priority:** SH | **Points:** 2

---

## Epic 2: Dashboard

### US-DASH-01 — Platform KPI Overview
**As a** Super Admin,
**I want to** see Total Tenants, Active Tenants, Expired Subscriptions, Total Admin Users, and Active Plans on the dashboard,
**So that** I can instantly assess overall platform health when I log in.

**Role:** Super Admin, Manager | **Priority:** MH | **Points:** 5

---

### US-DASH-02 — Subscription Trend Chart
**As a** Super Admin or Billing Admin,
**I want to** view a time-series chart of subscription counts over 6 or 12 months,
**So that** I can spot growth trends, churn spikes, and seasonal patterns.

**Role:** Super Admin, Billing Admin | **Priority:** SH | **Points:** 5

---

### US-DASH-03 — Live Activity Feed
**As a** Super Admin,
**I want to** see a continuously-updating feed of recent administrative actions with who did what, when, and the outcome,
**So that** I can monitor platform activity in near real-time and catch unexpected changes.

**Role:** Super Admin | **Priority:** MH | **Points:** 5

---

### US-DASH-04 — Quick Action Shortcuts
**As a** Super Admin,
**I want to** click tiles on the dashboard to jump directly to Create Tenant, Create Plan, Invite Admin, or Manage Policies,
**So that** I can reach my most frequent tasks in one click without navigating the sidebar.

**Role:** Super Admin | **Priority:** SH | **Points:** 2

---

### US-DASH-05 — Export Dashboard Report
**As a** Super Admin or Auditor,
**I want to** export a snapshot of dashboard KPIs as PDF or CSV,
**So that** I can include platform health data in weekly reports and stakeholder reviews.

**Role:** Super Admin, Auditor | **Priority:** CH | **Points:** 3

---

## Epic 3: Tenant Management

### US-TM-01 — View Tenant List
**As a** Super Admin or Manager,
**I want to** view a paginated, sortable list of all tenants with their name, code, plan, status, owner email, created date, and expiry,
**So that** I can quickly find any tenant and understand their current state at a glance.

**Role:** Super Admin, Manager, Support Lead | **Priority:** MH | **Points:** 5

---

### US-TM-02 — Filter Tenants
**As a** Super Admin or Manager,
**I want to** filter the tenant list by status, plan type, and creation date range,
**So that** I can quickly surface specific tenant segments (e.g., all Suspended Enterprise tenants).

**Role:** Super Admin, Manager, Support Lead | **Priority:** MH | **Points:** 3

---

### US-TM-03 — View Tenant Aggregate Stats
**As a** Super Admin or Billing Admin,
**I want to** see aggregate counts for Active Tenants, Total MRR, Suspended, and tenants expiring within 30 days at the bottom of the tenant list,
**So that** I can gauge revenue health and flag time-sensitive renewals without opening individual records.

**Role:** Super Admin, Billing Admin | **Priority:** SH | **Points:** 3

---

### US-TM-04 — Create New Tenant
**As a** Super Admin,
**I want to** provision a new client organization by filling in tenant details, owner info, admin credentials, subscription plan, and settings in a single form,
**So that** I can onboard a new customer onto the HROS platform quickly and completely.

**Role:** Super Admin | **Priority:** MH | **Points:** 8

---

### US-TM-05 — Configure Trial on Tenant Creation
**As a** Super Admin,
**I want to** optionally enable a trial period with an end date when creating a new tenant,
**So that** prospects can evaluate the platform before a paid subscription begins.

**Role:** Super Admin | **Priority:** MH | **Points:** 3

---

### US-TM-06 — Update Tenant Details
**As a** Super Admin or Support Lead,
**I want to** edit an existing tenant's organization info, owner contact, admin credentials, subscription plan, and settings,
**So that** I can keep tenant records accurate as their business details change.

**Role:** Super Admin, Support Lead | **Priority:** MH | **Points:** 5

---

### US-TM-07 — Archive Tenant
**As a** Super Admin,
**I want to** archive a terminated tenant with a confirmation step and a reason,
**So that** a churned client's access is revoked and their data is safely preserved for the retention period.

**Role:** Super Admin only | **Priority:** MH | **Points:** 5

---

### US-TM-08 — Search Tenants
**As any** admin user,
**I want to** search for a tenant by name, code, or owner email,
**So that** I can navigate to a specific tenant record in seconds regardless of the total tenant count.

**Role:** All roles | **Priority:** MH | **Points:** 3

---

## Epic 4: Subscription Management

### US-SUB-01 — View Subscription Detail
**As a** Super Admin or Billing Admin,
**I want to** view a tenant's full subscription page showing plan info, resource usage meters, payment method, and billing history,
**So that** I have a complete picture of a tenant's financial and capacity status in one view.

**Role:** Super Admin, Billing Admin, Support Lead | **Priority:** MH | **Points:** 5

---

### US-SUB-02 — Upgrade Subscription Plan
**As a** Super Admin or Billing Admin,
**I want to** upgrade a tenant to a higher plan with a prorated charge preview before confirming,
**So that** growing tenants can access more features and capacity without manual billing math.

**Role:** Super Admin, Billing Admin | **Priority:** MH | **Points:** 8

---

### US-SUB-03 — Downgrade Subscription Plan
**As a** Super Admin or Billing Admin,
**I want to** downgrade a tenant to a lower plan with usage validation to prevent mismatches,
**So that** cost-saving tenants can move to smaller plans only when their usage is safely within the new limits.

**Role:** Super Admin, Billing Admin | **Priority:** MH | **Points:** 8

---

### US-SUB-04 — Override Tenant Quotas
**As a** Super Admin,
**I want to** manually override a tenant's employee, admin, storage, and API call limits above their plan defaults,
**So that** I can honor custom agreements without creating a custom plan for each tenant.

**Role:** Super Admin | **Priority:** MH | **Points:** 5

---

### US-SUB-05 — Enable Subscription Trial
**As a** Super Admin or Billing Admin,
**I want to** enable a trial period for a tenant with a specified end date,
**So that** prospects can evaluate the product at no charge with defined boundaries.

**Role:** Super Admin, Billing Admin | **Priority:** MH | **Points:** 3

---

### US-SUB-06 — Extend Subscription Trial
**As a** Super Admin,
**I want to** extend a tenant's active trial by setting a new end date,
**So that** I can give high-potential prospects more evaluation time when needed.

**Role:** Super Admin | **Priority:** SH | **Points:** 2

---

### US-SUB-07 — Pause Subscription
**As a** Super Admin,
**I want to** temporarily pause a tenant's subscription,
**So that** a tenant on temporary hold stops being billed while their data is preserved.

**Role:** Super Admin | **Priority:** SH | **Points:** 5

---

### US-SUB-08 — Cancel Subscription
**As a** Super Admin,
**I want to** cancel a subscription effective at the end of the current billing period, with a confirmation dialog,
**So that** churned customers are cleanly offboarded at the right billing boundary.

**Role:** Super Admin | **Priority:** MH | **Points:** 5

---

### US-SUB-09 — Toggle Auto-Renew
**As a** Super Admin or Billing Admin,
**I want to** enable or disable automatic subscription renewal per tenant,
**So that** subscriptions either self-manage via auto-billing or require manual renewal attention.

**Role:** Super Admin, Billing Admin | **Priority:** MH | **Points:** 2

---

### US-SUB-10 — Toggle Billing Cycle
**As a** Super Admin or Billing Admin,
**I want to** switch a tenant's billing between monthly and yearly billing cycles,
**So that** I can align billing to tenant preferences and incentivize annual commitment.

**Role:** Super Admin, Billing Admin | **Priority:** SH | **Points:** 3

---

### US-SUB-11 — Add Internal Admin Notes
**As a** Super Admin,
**I want to** add private subscription notes visible only to super admins,
**So that** I can record negotiated rates and special arrangements without surfacing them to other roles.

**Role:** Super Admin | **Priority:** CH | **Points:** 2

---

### US-SUB-12 — Download All Invoices
**As a** Super Admin or Billing Admin,
**I want to** bulk-download all invoices for a tenant,
**So that** I can provide complete billing documentation to the tenant or finance team.

**Role:** Super Admin, Billing Admin | **Priority:** SH | **Points:** 3

---

## Epic 5: Plan Management

### US-PM-01 — View Plan Catalog
**As a** Super Admin or Billing Admin,
**I want to** see all subscription plans as side-by-side comparison cards with pricing, limits, and features,
**So that** I can review the current catalog and identify gaps or necessary updates.

**Role:** Super Admin, Billing Admin | **Priority:** MH | **Points:** 3

---

### US-PM-02 — Create Subscription Plan
**As a** Super Admin,
**I want to** define a new subscription plan with name, code, pricing, resource limits, feature permissions, and administrative settings,
**So that** I can expand the product catalog with new tiers for different customer segments.

**Role:** Super Admin | **Priority:** MH | **Points:** 8

---

### US-PM-03 — Edit Plan (Full Page)
**As a** Super Admin,
**I want to** modify an existing plan's full configuration on a dedicated edit page with an impact warning,
**So that** I can update the catalog confidently, knowing which tenants will be affected.

**Role:** Super Admin | **Priority:** MH | **Points:** 8

---

### US-PM-04 — Quick Edit Plan (Inline Panel)
**As a** Super Admin,
**I want to** make minor plan changes (price, limits, module toggles) from a slide-in panel on the plan list,
**So that** I don't need to navigate away from the catalog overview for small adjustments.

**Role:** Super Admin | **Priority:** SH | **Points:** 3

---

### US-PM-05 — Archive Plan
**As a** Super Admin,
**I want to** archive a plan to prevent new subscriptions to it without affecting existing subscribers,
**So that** I can retire obsolete tiers cleanly.

**Role:** Super Admin | **Priority:** MH | **Points:** 3

---

### US-PM-06 — Toggle Plan Visibility
**As a** Super Admin,
**I want to** toggle a plan's signup visibility independently of archiving,
**So that** I can temporarily hide plans under revision without fully retiring them.

**Role:** Super Admin | **Priority:** SH | **Points:** 2

---

## Epic 6: Admin Management

### US-AM-01 — View Admin List
**As a** Super Admin,
**I want to** view all internal admin accounts with their name, role, status, last login, and join date,
**So that** I can oversee who has portal access and spot stale or suspicious accounts.

**Role:** Super Admin | **Priority:** MH | **Points:** 3

---

### US-AM-02 — Invite Administrator
**As a** Super Admin,
**I want to** invite a new internal admin by entering their name, work email, and role, then sending them an invite link,
**So that** new HROS team members can be onboarded with appropriate portal access quickly.

**Role:** Super Admin | **Priority:** MH | **Points:** 5

---

### US-AM-03 — Assign / Change Admin Role
**As a** Super Admin,
**I want to** change any admin's role assignment,
**So that** I can align portal permissions to team members' actual responsibilities.

**Role:** Super Admin | **Priority:** MH | **Points:** 3

---

### US-AM-04 — Activate / Deactivate Admin
**As a** Super Admin,
**I want to** deactivate an admin's account (or reactivate a previously deactivated one),
**So that** I can immediately revoke or restore portal access as team membership changes.

**Role:** Super Admin | **Priority:** MH | **Points:** 3

---

### US-AM-05 — Reset Admin Password
**As a** Super Admin,
**I want to** trigger a password reset for any admin account,
**So that** I can resolve lockouts or respond to suspected credential compromise.

**Role:** Super Admin | **Priority:** SH | **Points:** 2

---

## Epic 7: Policy Management

### US-POL-01 — View Permission Matrix
**As a** Super Admin,
**I want to** see a visual matrix of View/Create/Update/Delete/Approve/Export permissions for each module and role,
**So that** I can review the current access control posture across the entire portal.

**Role:** Super Admin | **Priority:** MH | **Points:** 5

---

### US-POL-02 — Edit Role Permissions
**As a** Super Admin,
**I want to** check and uncheck permissions in the matrix per role and module, then save,
**So that** I can fine-tune access controls as team responsibilities and security requirements evolve.

**Role:** Super Admin | **Priority:** MH | **Points:** 5

---

### US-POL-03 — Build Policy Conditions
**As a** Super Admin,
**I want to** define contextual access conditions (e.g., deny CSV export for non-EU requests under GDPR),
**So that** I can enforce nuanced, rule-based access logic beyond flat role permissions.

**Role:** Super Admin | **Priority:** MH | **Points:** 8

---

### US-POL-04 — View Conflict Detection & Security Score
**As a** Super Admin,
**I want to** see a live security score and a list of detected permission conflicts after saving policies,
**So that** I can immediately identify and fix privilege overlaps or gaps before they become vulnerabilities.

**Role:** Super Admin | **Priority:** MH | **Points:** 5

---

### US-POL-05 — Create Custom Role
**As a** Super Admin,
**I want to** define a new named role with a blank permission set, then assign permissions via the matrix,
**So that** I can tailor access profiles for specialized team functions not covered by built-in roles.

**Role:** Super Admin | **Priority:** SH | **Points:** 5

---

### US-POL-06 — Run System Health Check
**As a** Super Admin,
**I want to** manually trigger a policy health check at any time,
**So that** I can validate security posture on-demand, not just after policy saves.

**Role:** Super Admin | **Priority:** SH | **Points:** 3

---

## Epic 8: Audit & Compliance

### US-AUD-01 — View Audit Logs
**As a** Super Admin or Auditor,
**I want to** browse a filterable, paginated audit log of all administrative events with entity, action, actor, timestamp, and before/after state,
**So that** I can investigate incidents and demonstrate compliance accountability.

**Role:** Super Admin, Auditor, Audit Specialist | **Priority:** MH | **Points:** 5

---

### US-AUD-02 — Export Audit Logs
**As an** Auditor or Audit Specialist,
**I want to** export the audit log (filtered or full) as CSV or PDF,
**So that** I can submit records for external compliance audits or long-term archival.

**Role:** Super Admin, Auditor, Audit Specialist | **Priority:** MH | **Points:** 3

---

### US-AUD-03 — View Policy-Filtered Audit Log
**As a** Super Admin,
**I want to** access audit logs pre-filtered to policy management events from within the Policies module,
**So that** I can review the history of permission changes without manually filtering the full log.

**Role:** Super Admin | **Priority:** SH | **Points:** 2

---

## Story Point Summary

| Epic | # Stories | Total Points | Priority Distribution |
|------|----------|--------------|-----------------------|
| Authentication | 6 | 18 | 4 MH, 2 SH |
| Dashboard | 5 | 20 | 3 MH, 1 SH, 1 CH |
| Tenant Management | 8 | 35 | 7 MH, 1 SH |
| Subscription Management | 12 | 52 | 8 MH, 3 SH, 1 CH |
| Plan Management | 6 | 27 | 4 MH, 2 SH |
| Admin Management | 5 | 16 | 4 MH, 1 SH |
| Policy Management | 6 | 31 | 4 MH, 2 SH |
| Audit & Compliance | 3 | 10 | 2 MH, 1 SH |
| **TOTAL** | **51** | **209** | **36 MH / 13 SH / 2 CH** |

*End of User Stories | HROS-US-001 v1.0*