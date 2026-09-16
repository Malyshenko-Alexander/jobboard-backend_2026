CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS vacancies (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employer_id      UUID NOT NULL,
    title            VARCHAR(255) NOT NULL,
    description      TEXT NOT NULL,
    requirements     TEXT NOT NULL,
    industry         VARCHAR(100) NOT NULL,
    salary_from      INT NULL,
    salary_to        INT NULL,
    experience_years INT NOT NULL DEFAULT 0,
    city             VARCHAR(100) NULL,
    is_active        BOOLEAN NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_vacancies_industry ON vacancies (industry);
CREATE INDEX IF NOT EXISTS idx_vacancies_salary_from ON vacancies (salary_from);
CREATE INDEX IF NOT EXISTS idx_vacancies_salary_to ON vacancies (salary_to);
CREATE INDEX IF NOT EXISTS idx_vacancies_experience ON vacancies (experience_years);
CREATE INDEX IF NOT EXISTS idx_vacancies_is_active ON vacancies (is_active);
CREATE INDEX IF NOT EXISTS idx_vacancies_employer_id ON vacancies (employer_id);

-- Optional snapshot filled by employer.updated events (stub for now).
CREATE TABLE IF NOT EXISTS company_snapshots (
    employer_id  UUID PRIMARY KEY,
    company_name VARCHAR(255) NOT NULL DEFAULT '',
    city         VARCHAR(100) NULL,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
