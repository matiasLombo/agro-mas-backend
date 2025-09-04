-- Add CBU and banking fields to users table for seller payments
ALTER TABLE users ADD COLUMN cbu VARCHAR(22);
ALTER TABLE users ADD COLUMN cbu_alias VARCHAR(100);
ALTER TABLE users ADD COLUMN bank_name VARCHAR(100);

-- Add comments for clarity
COMMENT ON COLUMN users.cbu IS 'CBU/CVU for receiving payments (22 digits)';
COMMENT ON COLUMN users.cbu_alias IS 'User-friendly alias for the CBU/CVU';
COMMENT ON COLUMN users.bank_name IS 'Name of the bank associated with the CBU/CVU';

-- Create index for CBU lookups (unique if provided)
CREATE UNIQUE INDEX idx_users_cbu ON users(cbu) WHERE cbu IS NOT NULL;