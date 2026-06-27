-- Start transaction
BEGIN;

-- Remove SSO mapping columns from admin_users table
ALTER TABLE admin_users DROP COLUMN IF EXISTS sso_identity_id;
ALTER TABLE admin_users DROP COLUMN IF EXISTS sso_provider;

COMMIT;
