-- Placeholder. The chat feature that originally lived here was reverted in 069650d.
-- This file exists so golang-migrate's versionExists(12) check passes on databases
-- where the original migration 12 was already applied. Actual cleanup of the
-- orphaned channels/messages tables is done in migration 000013.
SELECT 1;
