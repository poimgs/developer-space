-- Restore capacity to space_sessions
ALTER TABLE space_sessions ADD COLUMN capacity INTEGER NOT NULL DEFAULT 10;
ALTER TABLE space_sessions ADD CONSTRAINT space_sessions_capacity_positive CHECK (capacity > 0);

-- Restore capacity to session_series
ALTER TABLE session_series ADD COLUMN capacity INTEGER NOT NULL DEFAULT 10;
ALTER TABLE session_series ADD CONSTRAINT session_series_capacity_check CHECK (capacity > 0);

-- Drop image_url and location from session_series
ALTER TABLE session_series DROP COLUMN location;
ALTER TABLE session_series DROP COLUMN image_url;
