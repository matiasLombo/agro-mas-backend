-- Remove indexes for department and settlement
DROP INDEX IF EXISTS idx_products_location_hierarchy;
DROP INDEX IF EXISTS idx_products_settlement;
DROP INDEX IF EXISTS idx_products_department;

-- Remove department and settlement columns from products table
ALTER TABLE products DROP COLUMN IF EXISTS settlement;
ALTER TABLE products DROP COLUMN IF EXISTS department;