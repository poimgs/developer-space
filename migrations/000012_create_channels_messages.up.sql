CREATE TABLE channels (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    type        VARCHAR(20) NOT NULL DEFAULT 'general',
    session_id  UUID REFERENCES space_sessions(id) ON DELETE CASCADE,
    created_by  UUID NOT NULL REFERENCES members(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT channels_session_unique UNIQUE (session_id),
    CONSTRAINT channels_type_session CHECK (
        (type = 'session' AND session_id IS NOT NULL) OR
        (type = 'general' AND session_id IS NULL)
    )
);

CREATE INDEX idx_channels_type ON channels(type);
CREATE INDEX idx_channels_session_id ON channels(session_id) WHERE session_id IS NOT NULL;

CREATE TABLE messages (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id  UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    member_id   UUID NOT NULL REFERENCES members(id),
    content     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT messages_content_not_empty CHECK (length(trim(content)) > 0)
);

CREATE INDEX idx_messages_channel_created ON messages(channel_id, created_at DESC);
CREATE INDEX idx_messages_member ON messages(member_id);

-- Seed a default #general channel using the first admin member as creator
INSERT INTO channels (name, type, created_by)
SELECT 'general', 'general', id FROM members WHERE is_admin = true ORDER BY created_at ASC LIMIT 1;
