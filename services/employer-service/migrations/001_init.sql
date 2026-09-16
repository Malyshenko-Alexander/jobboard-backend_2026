CREATE TABLE IF NOT EXISTS employer_profiles (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL UNIQUE,
    company_name VARCHAR(255) NOT NULL DEFAULT '',
    description  TEXT NULL,
    website      VARCHAR(255) NULL,
    city         VARCHAR(100) NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
