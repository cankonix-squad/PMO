-- 000002_create_users_auth.up.sql
-- Creates users, user_sessions, and password_reset_tokens tables

CREATE TABLE IF NOT EXISTS users (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id  UUID         NOT NULL REFERENCES organizations(id),
    org_unit_id      UUID         REFERENCES org_units(id),
    employee_id      VARCHAR(100) UNIQUE,
    first_name       VARCHAR(100) NOT NULL,
    last_name        VARCHAR(100) NOT NULL DEFAULT '',
    email            VARCHAR(200) NOT NULL UNIQUE,
    password_hash    VARCHAR(255) NOT NULL,
    phone            VARCHAR(50),
    job_title        VARCHAR(200),
    avatar_url       TEXT,
    is_active        BOOLEAN      NOT NULL DEFAULT TRUE,
    must_change_pwd  BOOLEAN      NOT NULL DEFAULT TRUE,
    last_login_at    TIMESTAMPTZ,
    login_failed     INT          NOT NULL DEFAULT 0,
    locked_until     TIMESTAMPTZ,
    created_by       UUID         REFERENCES users(id),
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_users_organization_id ON users(organization_id);
CREATE INDEX IF NOT EXISTS idx_users_email           ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at      ON users(deleted_at);

CREATE TABLE IF NOT EXISTS user_sessions (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID         NOT NULL REFERENCES users(id),
    jti          VARCHAR(255) NOT NULL UNIQUE,
    refresh_jti  VARCHAR(255) NOT NULL UNIQUE,
    user_agent   TEXT,
    ip_address   VARCHAR(50),
    is_revoked   BOOLEAN      NOT NULL DEFAULT FALSE,
    expires_at   TIMESTAMPTZ  NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_jti     ON user_sessions(jti);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID         NOT NULL REFERENCES users(id),
    token_hash  VARCHAR(255) NOT NULL,
    expires_at  TIMESTAMPTZ  NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id    ON password_reset_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_token_hash ON password_reset_tokens(token_hash);
