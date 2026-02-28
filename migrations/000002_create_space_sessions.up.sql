CREATE TABLE space_sessions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    date        DATE NOT NULL,
    start_time  TIME NOT NULL,
    end_time    TIME NOT NULL,
    capacity    INTEGER NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'scheduled',
    created_by  UUID NOT NULL REFERENCES members(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT space_sessions_end_after_start CHECK (end_time > start_time),
    CONSTRAINT space_sessions_capacity_positive CHECK (capacity > 0),
    CONSTRAINT space_sessions_status_valid CHECK (status IN ('scheduled', 'shifted', 'canceled'))
);
