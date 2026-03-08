ALTER TABLE space_sessions ADD COLUMN capacity INTEGER CHECK (capacity > 0);
ALTER TABLE session_series ADD COLUMN capacity INTEGER CHECK (capacity > 0);
