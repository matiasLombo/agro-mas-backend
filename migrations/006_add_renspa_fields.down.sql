-- Rollback de campos RENSPA
DROP INDEX IF EXISTS idx_users_renspa;
ALTER TABLE users DROP COLUMN IF EXISTS establishment_location;
ALTER TABLE users DROP COLUMN IF EXISTS establishment_name;
ALTER TABLE users DROP COLUMN IF EXISTS renspa;