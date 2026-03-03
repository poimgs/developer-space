ALTER TABLE space_sessions
  DROP COLUMN IF EXISTS image_url,
  DROP COLUMN IF EXISTS location;
