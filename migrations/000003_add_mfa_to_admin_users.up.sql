-- Start transaction
BEGIN;

-- Add new MFA storage columns to admin_users table
ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS totp_secret VARCHAR(255);
ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS webauthn_credentials JSONB;

-- Migrate existing TOTP secret data from mfa_secret to totp_secret
UPDATE admin_users SET totp_secret = mfa_secret WHERE mfa_secret IS NOT NULL;

-- Drop the old mfa_secret column
ALTER TABLE admin_users DROP COLUMN IF EXISTS mfa_secret;

COMMIT;
