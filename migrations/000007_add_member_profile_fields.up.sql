ALTER TABLE members
  ADD COLUMN bio TEXT,
  ADD COLUMN skills TEXT[] DEFAULT '{}',
  ADD COLUMN linkedin_url VARCHAR(255),
  ADD COLUMN instagram_handle VARCHAR(255),
  ADD COLUMN github_username VARCHAR(255);
