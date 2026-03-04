-- Add image_url and location to session_series (series-level template fields)
ALTER TABLE session_series ADD COLUMN image_url VARCHAR(512);
ALTER TABLE session_series ADD COLUMN location TEXT;

-- Drop capacity from session_series
ALTER TABLE session_series DROP CONSTRAINT session_series_capacity_check;
ALTER TABLE session_series DROP COLUMN capacity;

-- Drop capacity from space_sessions
ALTER TABLE space_sessions DROP CONSTRAINT space_sessions_capacity_positive;
ALTER TABLE space_sessions DROP COLUMN capacity;
