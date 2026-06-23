-- Start transaction
BEGIN;

-- Restore the old mfa_secret column
ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS mfa_secret VARCHAR(255);

-- Copy data back from totp_secret to mfa_secret
UPDATE admin_users SET mfa_secret = totp_secret WHERE totp_secret IS NOT NULL;

-- Remove the new MFA storage columns
ALTER TABLE admin_users DROP COLUMN IF EXISTS totp_secret;
ALTER TABLE admin_users DROP COLUMN IF EXISTS webauthn_credentials;

COMMIT;
