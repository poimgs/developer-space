CREATE TABLE rsvps (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id  UUID NOT NULL REFERENCES space_sessions(id),
    member_id   UUID NOT NULL REFERENCES members(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT rsvps_unique_member_session UNIQUE (session_id, member_id)
);
