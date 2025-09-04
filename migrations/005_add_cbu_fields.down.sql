-- Remove CBU and banking fields from users table
DROP INDEX IF EXISTS idx_users_cbu;
ALTER TABLE users DROP COLUMN IF EXISTS bank_name;
ALTER TABLE users DROP COLUMN IF EXISTS cbu_alias;
ALTER TABLE users DROP COLUMN IF EXISTS cbu;