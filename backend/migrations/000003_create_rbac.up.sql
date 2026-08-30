-- 000003_create_rbac.up.sql
-- Creates roles, permissions, role_permissions, groups, user_roles, user_scopes

CREATE TABLE IF NOT EXISTS roles (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID         NOT NULL REFERENCES organizations(id),
    code            VARCHAR(100) NOT NULL,
    name            VARCHAR(200) NOT NULL,
    description     TEXT,
    is_system       BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    UNIQUE (organization_id, code)
);

CREATE INDEX IF NOT EXISTS idx_roles_organization_id ON roles(organization_id);
CREATE INDEX IF NOT EXISTS idx_roles_deleted_at      ON roles(deleted_at);

CREATE TABLE IF NOT EXISTS permissions (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    resource    VARCHAR(100) NOT NULL,
    action      VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (resource, action)
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS groups (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID         NOT NULL REFERENCES organizations(id),
    name            VARCHAR(200) NOT NULL,
    description     TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS group_roles (
    group_id   UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    role_id    UUID NOT NULL REFERENCES roles(id)  ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, role_id)
);

CREATE TABLE IF NOT EXISTS user_groups (
    user_id    UUID NOT NULL REFERENCES users(id)   ON DELETE CASCADE,
    group_id   UUID NOT NULL REFERENCES groups(id)  ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, group_id)
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id    UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    role_id    UUID NOT NULL REFERENCES roles(id)  ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS user_scopes (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID        NOT NULL REFERENCES organizations(id),
    scope_type      VARCHAR(50) NOT NULL,
    org_unit_id     UUID        REFERENCES org_units(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_scopes_user_id ON user_scopes(user_id);
