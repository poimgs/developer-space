ALTER TABLE members
  DROP COLUMN IF EXISTS bio,
  DROP COLUMN IF EXISTS skills,
  DROP COLUMN IF EXISTS linkedin_url,
  DROP COLUMN IF EXISTS instagram_handle,
  DROP COLUMN IF EXISTS github_username;
