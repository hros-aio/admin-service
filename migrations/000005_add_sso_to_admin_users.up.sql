-- Start transaction
BEGIN;

-- Add SSO mapping columns to admin_users table
ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS sso_identity_id VARCHAR(255) UNIQUE;
ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS sso_provider VARCHAR(255);

COMMIT;
