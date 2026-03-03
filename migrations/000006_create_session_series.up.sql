CREATE TABLE session_series (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    day_of_week INTEGER NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    start_time  TIME NOT NULL,
    end_time    TIME NOT NULL,
    capacity    INTEGER NOT NULL CHECK (capacity > 0),
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_by  UUID NOT NULL REFERENCES members(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT session_series_end_after_start CHECK (end_time > start_time)
);

ALTER TABLE space_sessions ADD COLUMN series_id UUID REFERENCES session_series(id) ON DELETE SET NULL;
CREATE UNIQUE INDEX idx_sessions_series_date ON space_sessions (series_id, date) WHERE series_id IS NOT NULL;
