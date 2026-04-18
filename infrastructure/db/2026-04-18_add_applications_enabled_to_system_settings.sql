ALTER TABLE system_settings
ADD COLUMN IF NOT EXISTS applications_enabled BOOLEAN NOT NULL DEFAULT TRUE;

UPDATE system_settings
SET applications_enabled = TRUE
WHERE applications_enabled IS DISTINCT FROM TRUE;
