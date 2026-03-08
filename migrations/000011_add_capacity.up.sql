ALTER TABLE space_sessions ADD COLUMN capacity INTEGER NOT NULL DEFAULT 20 CHECK (capacity > 0);
ALTER TABLE session_series ADD COLUMN capacity INTEGER NOT NULL DEFAULT 20 CHECK (capacity > 0);
