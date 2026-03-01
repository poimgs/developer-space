ALTER TABLE magic_tokens DROP CONSTRAINT magic_tokens_member_id_fkey,
  ADD CONSTRAINT magic_tokens_member_id_fkey FOREIGN KEY (member_id) REFERENCES members(id);

ALTER TABLE space_sessions DROP CONSTRAINT space_sessions_created_by_fkey,
  ADD CONSTRAINT space_sessions_created_by_fkey FOREIGN KEY (created_by) REFERENCES members(id);

ALTER TABLE rsvps DROP CONSTRAINT rsvps_session_id_fkey,
  ADD CONSTRAINT rsvps_session_id_fkey FOREIGN KEY (session_id) REFERENCES space_sessions(id);
