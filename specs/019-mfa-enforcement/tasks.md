# Tasks: MFA Enforcement (Super Admins) — Migration (TSK-MFA-001)

**Input**: Design documents from `/specs/019-mfa-enforcement/`

**Prerequisites**: plan.md ✅, spec.md ✅

**Scope**: Create up/down SQL migration scripts to add MFA credential storage to the `admin_users` table.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on each other)
- **[Story]**: Maps to user story in spec.md
- Include exact file paths in every task description

---

## Phase 1: Migration (TSK-MFA-001) 🔲 Pending

- [x] T001 [US1] Create up migration script at `migrations/000003_add_mfa_to_admin_users.up.sql` to add `totp_secret` (VARCHAR) and `webauthn_credentials` (JSONB) columns to the `admin_users` table, and migrate any existing data from `mfa_secret` to `totp_secret`, before dropping `mfa_secret`.
- [x] T002 [US1] Create down migration script at `migrations/000003_add_mfa_to_admin_users.down.sql` to revert the migration by recreating the `mfa_secret` (VARCHAR) column, restoring its values from `totp_secret`, and dropping `totp_secret` and `webauthn_credentials` columns.
