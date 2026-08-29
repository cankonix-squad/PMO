-- 000037_notifications.up.sql
-- Notification outbox/inbox table for UAT-004 Notification Delivery Foundation.
-- Stores both IN_APP and EMAIL notification records with delivery status.

CREATE TABLE IF NOT EXISTS notifications (
    id                UUID         NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    organization_id   UUID         NOT NULL REFERENCES organizations(id),
    recipient_user_id UUID         REFERENCES users(id),
    channel           VARCHAR(20)  NOT NULL DEFAULT 'IN_APP'
                          CHECK (channel IN ('IN_APP','EMAIL')),
    status            VARCHAR(20)  NOT NULL DEFAULT 'PENDING'
                          CHECK (status IN ('PENDING','SENT','FAILED','READ')),
    priority          VARCHAR(20)  NOT NULL DEFAULT 'NORMAL'
                          CHECK (priority IN ('LOW','NORMAL','HIGH','URGENT')),
    subject           VARCHAR(500) NOT NULL,
    body              TEXT         NOT NULL,
    source_type       VARCHAR(100),
    source_id         VARCHAR(255),
    error_message     TEXT,
    sent_at           TIMESTAMPTZ,
    read_at           TIMESTAMPTZ,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_notifications_org        ON notifications(organization_id);
CREATE INDEX IF NOT EXISTS idx_notifications_recipient  ON notifications(recipient_user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_status     ON notifications(status);
CREATE INDEX IF NOT EXISTS idx_notifications_channel    ON notifications(channel);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_deleted_at ON notifications(deleted_at);
