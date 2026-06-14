# Acceptance Criteria
## HROS Admin — Super Admin Portal
**Document ID:** HROS-AC-001 | **Version:** 1.0
**Format:** Given [context] / When [action] / Then [outcome]
> `[A]` = Assumption inferred from SaaS HRMS best practice

---

## AC-AUTH: Authentication

### AC-AUTH-01: Successful Login
**Story:** US-AUTH-01

**Scenario 1 — Valid credentials**
- **Given** I am on the /login page
- **When** I enter a registered admin email and correct password and click "Sign In to Dashboard"
- **Then** I am redirected to /dashboard within 3 seconds
- **And** the top bar shows my name and role label
- **And** an access token and refresh token are created for my session

**Scenario 2 — Invalid password**
- **Given** I am on the /login page
- **When** I enter a registered email and incorrect password
- **Then** a generic error message appears: "Invalid email or password"
- **And** the error does not specify which field is incorrect
- **And** I remain on the /login page

**Scenario 3 — Unregistered email**
- **Given** I am on the /login page
- **When** I enter an email not linked to any admin account
- **Then** the same generic error appears: "Invalid email or password"
- **And** the response time is the same as for an invalid password (no timing oracle)

**Scenario 4 — Blank fields**
- **Given** I am on the /login page
- **When** I click "Sign In to Dashboard" with email or password fields empty
- **Then** inline validation errors appear on the blank field(s)
- **And** no API request is made

---

### AC-AUTH-02: Account Lockout
**Story:** US-AUTH-06

- **Given** I am on the /login page
- **When** I submit incorrect credentials 5 consecutive times within 15 minutes `[A]`
- **Then** my account is locked for 30 minutes
- **And** I receive an email notification with the lock timestamp and unlock time
- **And** subsequent login attempts during the lock period display: "Account locked. Try again after [time] or contact IT Support."
- **And** successful login within the 5 attempts resets the counter

---

### AC-AUTH-03: Persistent Session
**Story:** US-AUTH-05

- **Given** I am on the /login page
- **When** I check "Keep me logged in for 30 days" and successfully log in
- **Then** my session persists for 30 days even if I close the browser
- **And** if "Keep me logged in" is unchecked, my session ends when the browser is closed

---

### AC-AUTH-04: Password Reset
**Story:** US-AUTH-04

**Scenario 1 — Successful reset**
- **Given** I am on /forgot-password
- **When** I enter my registered admin email and submit
- **Then** I see: "If an account exists for that email, a reset link has been sent"
- **And** I receive an email within 2 minutes `[A]` containing a single-use reset link
- **And** the link expires after 60 minutes

**Scenario 2 — Complete reset**
- **Given** I have received and clicked a valid reset link
- **When** I enter a valid new password (min 10 chars, 1 uppercase, 1 number, 1 special) `[A]` and confirm it
- **Then** my password is updated
- **And** all my existing active sessions are invalidated
- **And** I am redirected to /login with a success toast: "Password updated. Please log in."

**Scenario 3 — Expired link**
- **Given** I click a password reset link that is more than 60 minutes old
- **Then** I see: "This link has expired. Request a new password reset."

---

## AC-DASH: Dashboard

### AC-DASH-01: KPI Cards
**Story:** US-DASH-01

- **Given** I am logged in as any admin role
- **When** I navigate to /dashboard
- **Then** all 5 KPI cards (Total Tenants, Active Tenants, Expired Subs, Total Admin Users, Active Plans) display within 2 seconds
- **And** Total Tenants shows a +/- % trend vs. the prior 30-day period
- **And** Active Tenants shows a stability label ("Stable" / "Growing" / "Declining") `[A]`
- **And** data is refreshed automatically every 60 seconds without a page reload

---

### AC-DASH-02: Activity Feed
**Story:** US-DASH-03

- **Given** I am on /dashboard
- **Then** the Recent Tenant Activity feed shows the 20 most recent admin events
- **And** each row shows: tenant name with avatar initials, action, status badge, date/time, and operator name
- **And** the feed auto-refreshes every 60 seconds
- **And** "View All Activities" navigates to a full audit log page

---

### AC-DASH-03: Trend Chart Toggle
**Story:** US-DASH-02

- **Given** I am on /dashboard
- **When** I click "6M" the chart shows the last 6 calendar months of subscription data
- **When** I click "1Y" the chart switches to the last 12 calendar months
- **And** hovering over a data point shows a tooltip with that month's Active, Trial, and Expired counts

---

## AC-TM: Tenant Management

### AC-TM-01: Tenant List and Filters
**Story:** US-TM-01, US-TM-02

- **Given** I navigate to /tenants
- **Then** I see a table with columns: Tenant Name, Code, Owner Email, Plan, Status, Created, Expiry
- **And** the first page shows 10 results by default
- **And** total tenant count and pagination controls are shown
- **And** the aggregate footer shows: Active Tenants, Total MRR, Suspended, Expiring (30d)

**Filter scenarios:**
- **When** I select "Suspended" from the Status filter → only suspended tenants appear
- **When** I select "Enterprise" from Plan Type filter → only Enterprise plan tenants appear
- **When** I set a Created Date range → only tenants created in that range appear
- **When** I click "Clear Filters" → all filters reset and full list reloads

---

### AC-TM-02: Create Tenant — Happy Path
**Story:** US-TM-04

- **Given** I am a Super Admin on /tenants/new
- **When** I fill all required fields across all 5 sections and click "Create Tenant"
- **Then** a success toast appears: "Tenant [name] created successfully"
- **And** I am navigated to the tenant detail page for the new tenant
- **And** the new tenant appears in the tenant list
- **And** the tenant's initial admin receives an onboarding email within 5 minutes `[A]`
- **And** an audit log entry is created for the 'tenant.created' event

**Required field validation:**
- **When** I submit without Tenant Name → inline error: "Tenant name is required"
- **When** I submit without Tenant Code → inline error: "Tenant code is required"
- **When** I enter a Tenant Code that already exists → error: "Tenant code already exists"
- **When** I enter an Admin Email already used → error: "Email is already associated with an admin account"
- **When** I enable the trial period and set Trial End Date before Start Date → error: "Trial end date must be after start date"

---

### AC-TM-03: Tenant Code Immutability
**Story:** US-TM-06

- **Given** I am editing an existing tenant on /tenants/{id}/edit
- **Then** the Tenant Code field is read-only and cannot be modified
- **And** no change to Tenant Code is accepted even via API `[A]`

---

### AC-TM-04: Archive Tenant
**Story:** US-TM-07

**Scenario 1 — Successful archive**
- **Given** I am a Super Admin on the Update Tenant page for an active tenant
- **When** I click "Archive Tenant"
- **Then** a confirmation modal appears requiring me to type the exact tenant name
- **And** the Confirm button is disabled until the name matches exactly
- **When** I type the correct name and click Confirm
- **Then** tenant status changes to "Archived"
- **And** all active subscriptions for that tenant are cancelled
- **And** I am navigated back to the tenant list
- **And** the archived tenant is no longer visible in the default tenant list `[A]`
- **And** an audit log entry records the archive event, reason, and acting admin

**Scenario 2 — Permission gate**
- **Given** I am logged in as a Manager
- **Then** the "Archive Tenant" button is not visible on the tenant edit page

---

### AC-TM-05: Update Tenant
**Story:** US-TM-06

- **Given** I am on the Update Tenant page for "Global Logistics Inc."
- **Then** all fields are pre-populated with current values (e.g., Tenant Code = "GL-8842" is read-only)
- **When** I change the Industry to "Finance" and click "Save Changes"
- **Then** a success toast appears: "Tenant updated successfully"
- **And** the audit log records: `{ field: "industry", old: "Logistics", new: "Finance" }`
- **And** unchanged fields are not written to the audit log

---

## AC-SUB: Subscription Management

### AC-SUB-01: Resource Usage Color Coding
**Story:** US-SUB-01

- **Given** I am on a tenant's Subscription Detail page
- **Then** employee usage bar is **blue** when usage < 70%
- **And** employee usage bar is **orange** when usage is 70–89%
- **And** employee usage bar is **red** when usage is ≥ 90%
- **And** Cloud Storage (46 GB / 50 GB = 92%) shows a red bar with a "Critical" indicator

---

### AC-SUB-02: Plan Upgrade with Proration
**Story:** US-SUB-02

- **Given** I am on the Manage Subscription page for a tenant on the Pro plan
- **When** I select "Enterprise" from the plan dropdown and select "Immediately" and click "Upgrade Plan"
- **Then** a confirmation modal appears showing:
  - Current plan → New plan
  - Prorated charge amount for the remaining billing days
  - Next full billing amount
- **When** I confirm
- **Then** the subscription plan updates to Enterprise immediately
- **And** a prorated invoice is generated
- **And** an audit log entry records the plan change

---

### AC-SUB-03: Downgrade Blocked by Usage
**Story:** US-SUB-03

- **Given** a tenant has 842 employees on an Enterprise plan (employee limit: 1,000)
- **When** I try to downgrade to Pro plan (employee limit: 500)
- **Then** the Downgrade button is blocked with an error:
  "Cannot downgrade: Current employee count (842) exceeds the Pro plan limit (500). Reduce headcount before downgrading."
- **And** no subscription change is made

---

### AC-SUB-04: Trial Enable and Extend
**Story:** US-SUB-05, US-SUB-06

**Enable Trial:**
- **Given** a tenant has never had a trial on their current plan
- **When** I click "Enable Trial" and set a trial end date 14 days out
- **Then** subscription status changes to "Trial"
- **And** a trial end warning email is scheduled for 3 days and 1 day before end date

**Re-enable blocked:**
- **Given** a tenant has already used a trial on their current plan
- **When** I view Trial Management
- **Then** "Enable Trial" button is disabled with tooltip: "Trial has already been used for this plan"

**Extend Trial:**
- **Given** a tenant is in active Trial status
- **When** I update the trial end date to a later date and click "Extend Trial"
- **Then** the trial_end_date is updated
- **And** warning emails are rescheduled to new end date

---

### AC-SUB-05: Unsaved Changes Protection
**Story:** US-SUB-01 (implicit)

- **Given** I have changed the billing cycle toggle on the Manage Subscription page without saving
- **Then** the "Unsaved Changes" sticky panel appears at bottom-right
- **When** I try to navigate away
- **Then** a browser dialog warns: "Leave page? Changes you made may not be saved."
- **When** I click "Discard Draft"
- **Then** the form resets to the last saved state and the unsaved panel disappears

---

### AC-SUB-06: Cancel Subscription
**Story:** US-SUB-08

- **Given** I am on a tenant's subscription management page
- **When** I click "Cancel Subscription"
- **Then** a confirmation dialog appears explaining: "Subscription will end on [billing_period_end]. Data will be retained for 90 days."
- **When** I confirm
- **Then** subscription status changes to "Cancelled"
- **And** an audit log entry is created
- **And** tenant continues to have access until the billing period end date

---

## AC-PM: Plan Management

### AC-PM-01: Create Plan — Validation
**Story:** US-PM-02

- **Given** I am on /plans/new
- **When** I submit without Plan Name → error: "Plan name is required"
- **When** I enter a Plan Code that already exists → error: "Plan code already exists"
- **When** I enter a negative Monthly Price → error: "Price must be 0 or greater"
- **When** I enter a Trial Period over 365 → error: "Trial period cannot exceed 365 days"
- **When** all required fields are valid and I click "Create Plan"
- **Then** the plan is created and visible in the plan catalog immediately
- **And** a success toast: "Plan [name] created successfully"
- **And** an audit log entry is created

---

### AC-PM-02: Plan Code Immutability
**Story:** US-PM-03

- **Given** I am editing an existing plan
- **Then** the Plan Code field is read-only
- **And** a label indicates: "Plan code cannot be changed after creation"

---

### AC-PM-03: Edit Plan with Impact Warning
**Story:** US-PM-03

- **Given** I change the monthly price of the "Pro Plan" which has 50 active subscribers
- **When** I click "Save Changes"
- **Then** a warning banner appears: "Pricing change will affect 50 active subscriptions at their next billing cycle."
- **And** I must click a secondary "Confirm Save" button in the warning to proceed
- **And** an audit log entry records the price change with old and new values

---

### AC-PM-04: Archive Plan
**Story:** US-PM-05

- **Given** the "Basic Plan" has 12 active tenant subscribers
- **When** I archive the Basic Plan
- **Then** the plan is no longer selectable in the Create Tenant or Change Plan forms
- **And** the 12 existing subscribers on Basic remain unchanged
- **And** the plan shows as "Archived" in the plan list (hidden from new signups)
- **And** if the Basic Plan is the only active plan → archive action is blocked with: "Cannot archive the only active plan"

---

## AC-AM: Admin Management

### AC-AM-01: Invite Admin — Happy Path
**Story:** US-AM-02

- **Given** I am a Super Admin on the Admins page with the Invite panel open
- **When** I enter Full Name, Work Email, select "Manager" role, leave status "Active", and click "Send Invitation"
- **Then** a toast appears: "Invitation sent to [email]"
- **And** the invitee receives an email with a link valid for 48 hours
- **And** the admin record appears in the list with status "Pending"
- **And** clicking the link navigates to a password-set page; completing it sets status to "Active"

**Validation:**
- **When** I enter an email already used by another admin → error: "An admin with this email already exists"
- **When** a non-Super Admin attempts to invite a Super Admin role → error: "Only Super Admins can invite Super Admins" (or role option hidden)

---

### AC-AM-02: Self-Protection
**Story:** US-AM-04

- **Given** I am logged in as "Alex Rivera" (Super Admin)
- **When** I open Alex Rivera's admin row actions menu
- **Then** "Deactivate" and "Remove Admin" options are disabled with tooltip: "You cannot modify your own account"

---

### AC-AM-03: Invitation Expiry
**Story:** US-AM-02

- **Given** an invitation link is more than 48 hours old
- **When** the invitee clicks the link
- **Then** they see: "This invitation has expired. Ask an administrator to resend the invitation."
- **And** the admin record status remains "Pending" `[A]`

---

## AC-POL: Policy Management

### AC-POL-01: Permission Matrix Save
**Story:** US-POL-02

- **Given** I am a Super Admin on the Policies page with the "Manager" role selected
- **When** I check the "Create" permission for Tenant Management and click "Save Changes"
- **Then** the permission is saved
- **And** the conflict detection runs
- **And** the security score updates
- **And** an audit log entry records the permission change with before/after snapshot

---

### AC-POL-02: Super Admin Matrix is Read-Only
**Story:** US-POL-01

- **Given** I select "Super Admin" in the System Roles panel
- **Then** all permission checkboxes are checked and visually disabled (read-only)
- **And** the "Save Changes" button is hidden or disabled for the Super Admin role

---

### AC-POL-03: Conflict Detection Output
**Story:** US-POL-04

- **Given** I have saved a permission configuration that creates a privilege overlap
- **Then** the Security Score decreases by the conflict deduction amount
- **And** a conflict card is displayed describing the conflict
- **And** a "Run System Health Check" link is available for detailed analysis

---

### AC-POL-04: Policy Condition — GDPR Rule
**Story:** US-POL-03

- **Given** I build a condition: Subject="Billing Admin", Action="Export CSV", Condition="location outside EU", Effect="Deny"
- **When** I click "Append New Logic Statement"
- **Then** the condition card appears in Active Logic Chains tagged "GDPR Compliance"
- **And** saving the policy results in this rule being evaluated before all other allow rules for that action
- **And** a Billing Admin from outside the EU attempting to export a CSV receives a 403 Forbidden

---

### AC-POL-05: Custom Role Creation
**Story:** US-POL-05

- **Given** I click "+ Add New Role"
- **When** I enter a role name "Regional Manager" and confirm
- **Then** the new role appears in the System Roles list with no permissions checked
- **And** I can assign permissions via the matrix and save

---

## AC-AUD: Audit Logging

### AC-AUD-01: Audit Log Completeness
**Story:** US-AUD-01

For each of the following actions, an audit log entry MUST be created:
- Tenant created / updated / archived
- Subscription plan changed / paused / cancelled
- Plan created / updated / archived
- Admin invited / role changed / deactivated / removed
- Policy permission changed / condition added
- Any CSV/PDF export initiated
- Login success / login failure / logout

Each entry MUST contain:
- event_type, entity_type, entity_id, entity_name
- operator_id, operator_name, operator_type
- ip_address, created_at (UTC)
- previous_state and new_state (JSON)

---

### AC-AUD-02: Audit Log Immutability
**Story:** US-AUD-01

- **Given** I am a Super Admin
- **When** I attempt to delete or modify an audit log entry via the UI or API
- **Then** the action is rejected with: "Audit logs are immutable and cannot be modified"
- **And** no UPDATE or DELETE SQL is executed against the audit_logs table

---

### AC-AUD-03: Export Audit Log
**Story:** US-AUD-02

- **Given** I am an Auditor or Super Admin on the audit log page
- **When** I apply filters (date range, event type) and click "Export"
- **Then** a download is triggered within 5 seconds for files < 10,000 rows `[A]`
- **And** the exported file (CSV or PDF) contains all columns from the filtered result set
- **And** the export action itself is recorded in the audit log

---

*End of Acceptance Criteria | HROS-AC-001 v1.0*