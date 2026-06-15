-- Initial migration for schema verification
CREATE TABLE IF NOT EXISTS schema_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO schema_verifications (status) VALUES ('INITIALIZED');
