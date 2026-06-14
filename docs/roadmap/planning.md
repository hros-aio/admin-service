Based on a comprehensive architectural and project management analysis of the uploaded system requirement specifications, user stories, API documentation, database models, permission matrices, and UI designs, here is the structured execution plan for the HROS Admin Super Portal.

# Product Roadmap

The delivery of the HROS Super Admin Portal is phased to ensure infrastructure and security prerequisites are established before complex multi-tenant billing logic is introduced.

*   **Phase 1: Foundation & Security (Months 1-2):** Database schema, Auth, Admin Management, Audit Logging, and RBAC Policy Engine.
*   **Phase 2: Core Provisioning (Months 2-3):** Plan Management and Tenant Management.
*   **Phase 3: Financial & Lifecycle (Months 3-4):** Subscription Management, Stripe Integration, Usage Metrics.
*   **Phase 4: Analytics & Optimization (Month 4):** Dashboard, KPI caching, Exports, UI/UX polish.

# Epic List

As defined in the User Stories document, the system is structured into 8 core Epics (totaling 51 User Stories and 209 points):

1.  **Epic 1:** Authentication & Session Management (18 pts)
2.  **Epic 2:** Dashboard (20 pts)
3.  **Epic 3:** Tenant Management (35 pts)
4.  **Epic 4:** Subscription Management (52 pts)
5.  **Epic 5:** Plan Management (27 pts)
6.  **Epic 6:** Admin Management (16 pts)
7.  **Epic 7:** Policy Management (31 pts)
8.  **Epic 8:** Audit & Compliance (10 pts)

# Feature Breakdown

### Epic 1: Authentication & Session Management
*   **Feature 1.1: Core Identity**
    *   *US-AUTH-01 (Credential Login)*
        *   Task 1: Build `admin_users` and `roles` DB schema [S]
        *   Task 2: Implement bcrypt hashing and JWT token issuance via `POST /auth/login` [M]
        *   Task 3: Build UI Login Page with masked password toggle [XS]
    *   *US-AUTH-04 (Forgot Password)*
        *   Task 1: Generate & store short-lived reset tokens [S]
        *   Task 2: Implement SES email dispatch [M]
*   **Feature 1.2: Session Security**
    *   *US-AUTH-06 (Account Lockout)*
        *   Task 1: Implement Redis counter for failed logins (5 attempts / 15 mins) [S]
    *   *US-AUTH-05 (Persistent Session)*
        *   Task 1: Implement 30-day refresh token logic via `POST /auth/refresh` [S]

### Epic 2: Dashboard
*   **Feature 2.1: Analytics Engine**
    *   *US-DASH-01 (Platform KPI Overview)*
        *   Task 1: Build Redis cache layer with 60-second TTL for aggregate queries [M]
        *   Task 2: Build UI metric cards with trend indicators [S]
    *   *US-DASH-02 (Subscription Trend Chart)*
        *   Task 1: Implement time-series data aggregation endpoint [M]
        *   Task 2: Integrate frontend charting library (6M/1Y toggle) [M]
*   **Feature 2.2: Live Monitoring**
    *   *US-DASH-03 (Activity Feed)*
        *   Task 1: API `GET /dashboard/activity` pulling top 20 logs [S]
        *   Task 2: UI feed with auto-polling (60s) and status badge rendering [S]

### Epic 3: Tenant Management
*   **Feature 3.1: Tenant Lifecycle**
    *   *US-TM-04 (Create New Tenant)*
        *   Task 1: API `POST /tenants` utilizing atomic DB transaction (Tenant + Owner + Sub + Admin) [XL]
        *   Task 2: Build 5-section UI Wizard [L]
    *   *US-TM-06 (Update Tenant Details)*
        *   Task 1: API `PUT /tenants/:id` with diff-based partial updates and immutable `tenant_code` rule [M]
    *   *US-TM-07 (Archive Tenant)*
        *   Task 1: Implement soft-delete cascade (cancel subs, invalidate sessions) [M]

### Epic 4: Subscription Management
*   **Feature 4.1: Financial Operations**
    *   *US-SUB-02 (Upgrade Plan)*
        *   Task 1: Stripe integration for mid-cycle proration math [XL]
        *   Task 2: UI confirmation modal with real-time charge preview [M]
    *   *US-SUB-03 (Downgrade Plan)*
        *   Task 1: Build usage validation service to block drops below current limits [L]
*   **Feature 4.2: Controls & Quotas**
    *   *US-SUB-04 (Override Quotas)*
        *   Task 1: API support for manual limit overrides [S]
    *   *US-SUB-07 (Pause/Cancel)*
        *   Task 1: Implement subscription state machine transitions [M]

### Epic 5: Plan Management
*   **Feature 5.1: Plan Cataloging**
    *   *US-PM-02 (Create Plan)*
        *   Task 1: `plans` and `plan_features` DB schema [S]
        *   Task 2: UI form for pricing, resource limits, and feature toggles [M]
    *   *US-PM-03 (Edit Plan)*
        *   Task 1: API `PUT /plans/:id` with downstream active subscriber warning logic [M]

### Epic 6: Admin Management
*   **Feature 6.1: Internal Access**
    *   *US-AM-02 (Invite Administrator)*
        *   Task 1: Create 48-hour secure invite tokens and state machine [M]
        *   Task 2: UI slide-in panel for invites [S]
    *   *US-AM-03 (Role Assignment)*
        *   Task 1: Super Admin restriction gate logic [S]

### Epic 7: Policy Management
*   **Feature 7.1: RBAC & Rules Engine**
    *   *US-POL-01 & 02 (Permission Matrix)*
        *   Task 1: UI grid matrix for View/Create/Update/Delete/Approve/Export [L]
        *   Task 2: Immutable Super Admin bypass logic [S]
    *   *US-POL-03 (Condition Builder)*
        *   Task 1: Runtime context evaluation engine (e.g., GDPR rules) [XL]
    *   *US-POL-04 (Conflict Detection)*
        *   Task 1: Algorithm to calculate Security Score and detect ALLOW/DENY overlaps [L]

### Epic 8: Audit & Compliance
*   **Feature 8.1: System Accountability**
    *   *US-AUD-01 (View Audit Logs)*
        *   Task 1: Database-level `RULE` to strictly enforce append-only immutability [M]
        *   Task 2: Global API middleware to intercept and log all state-changing operations [L]

# Dependency Graph

Tasks must follow a strict architectural sequence to prevent rework:

1.  **DB Schema Creation** (Admin, Roles, Tokens) → **Unblocks:** `Epic 1: Auth`
2.  **Auth & Session** → **Unblocks:** `Epic 6: Admin Mgmt` & `Epic 8: Audit Logs`
3.  **Audit System Foundation** (Must be in place before any entity mutations occur) → **Unblocks:** `All write operations`
4.  **Epic 7: Policy Mgmt (RBAC)** → **Unblocks:** `API Gateway routing & authorization`
5.  **Epic 5: Plan Mgmt** (Need plans to assign) → **Unblocks:** `Epic 3: Tenant Mgmt`
6.  **Epic 3: Tenant Mgmt** (Need tenants to bill) → **Unblocks:** `Epic 4: Subscription Mgmt`
7.  **Epic 4: Subscription Mgmt** → **Unblocks:** `Epic 2: Dashboard` (Relies on tenant/sub data)

# Implementation Order

1.  **Sprint 1:** DB Design, Global Audit Middleware, Core Login & JWT Auth.
2.  **Sprint 2:** Admin Management (Invite/Roles), RBAC Foundation (Matrix).
3.  **Sprint 3:** Policy Engine (Conditions, Conflict Detection), Plan Management (CRUD).
4.  **Sprint 4:** Tenant Creation Wizard (Atomic DB routines), Tenant Listing.
5.  **Sprint 5:** Stripe Integration, Subscription Upgrades/Downgrades, Usage Validation.
6.  **Sprint 6:** Subscription Operations (Pause/Cancel/Trial), Dashboard APIs & Caching.
7.  **Sprint 7:** UI Dashboard Implementation, Activity Feed, Export Reporting, Penetration Testing.

# Sprint Recommendations

*   **Sprint 1 Focus (Foundation):** Address the highest technical risk early—the **Global Audit Middleware**. All subsequent endpoints rely on this for append-only compliance.
*   **Sprint 3 Focus (Security):** The **Policy Condition Builder & Runtime Evaluation**. *Technical Risk:* Evaluating context-based logic (like GDPR checks) on every API request could violate the < 500ms read latency SLA. *Mitigation:* Implement heavy Redis caching for policy rules.
*   **Sprint 4 Focus (Architecture):** **Tenant Creation**. *Technical Risk:* Creating a tenant involves writing to 5 different tables. *Mitigation:* Must implement strict PostgreSQL atomic transactions to prevent orphaned records if a failure occurs mid-creation.
*   **Sprint 5 Focus (External Integration):** **Stripe & Prorations**. *Technical Risk:* Handling mid-cycle upgrades requires complex proration math. *Mitigation:* Rely entirely on Stripe's billing engine for the calculation; the portal should only manage metadata to prevent financial desync.