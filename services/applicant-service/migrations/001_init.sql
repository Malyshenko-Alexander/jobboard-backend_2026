CREATE TABLE IF NOT EXISTS applicant_profiles (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL UNIQUE,
    full_name  VARCHAR(255) NOT NULL DEFAULT '',
    phone      VARCHAR(50) NULL,
    city       VARCHAR(100) NULL,
    about      TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS resumes (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    applicant_id     UUID NOT NULL REFERENCES applicant_profiles(id) ON DELETE CASCADE,
    title            VARCHAR(255) NOT NULL DEFAULT '',
    summary          TEXT NULL,
    skills           TEXT NULL,
    experience_years INT NOT NULL DEFAULT 0,
    education        TEXT NULL,
    desired_salary   INT NULL,
    is_active        BOOLEAN NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_resumes_applicant_id ON resumes (applicant_id);
