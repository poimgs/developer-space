DROP INDEX IF EXISTS idx_sessions_series_date;
ALTER TABLE space_sessions DROP COLUMN IF EXISTS series_id;
DROP TABLE IF EXISTS session_series;
