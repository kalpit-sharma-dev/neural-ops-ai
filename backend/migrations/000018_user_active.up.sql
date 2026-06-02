-- User lifecycle: allow admins to deactivate users without deletion.

ALTER TABLE users ADD COLUMN IF NOT EXISTS active BOOLEAN NOT NULL DEFAULT TRUE;
