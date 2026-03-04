ALTER TABLE session_series
  ADD COLUMN every_n_weeks INTEGER NOT NULL DEFAULT 1
  CHECK (every_n_weeks BETWEEN 1 AND 4);
