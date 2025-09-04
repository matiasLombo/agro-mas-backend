-- Rollback location name fields from products table
-- This migration removes the province_name, department_name, and settlement_name columns
-- and their associated indexes

-- Drop indexes first
DROP INDEX IF EXISTS idx_products_location_name_hierarchy;
DROP INDEX IF EXISTS idx_products_settlement_name;
DROP INDEX IF EXISTS idx_products_department_name;
DROP INDEX IF EXISTS idx_products_province_name;

-- Drop columns
ALTER TABLE products DROP COLUMN IF EXISTS settlement_name;
ALTER TABLE products DROP COLUMN IF EXISTS department_name;
ALTER TABLE products DROP COLUMN IF EXISTS province_name;