-- Add location name fields to products table
-- These fields store the human-readable names for provinces, departments, and settlements
-- while the existing province/department/settlement fields store codes/IDs

ALTER TABLE products ADD COLUMN province_name VARCHAR(100);
ALTER TABLE products ADD COLUMN department_name VARCHAR(100);
ALTER TABLE products ADD COLUMN settlement_name VARCHAR(100);

-- Create indexes for the new location name columns for better query performance
CREATE INDEX idx_products_province_name ON products(province_name) WHERE province_name IS NOT NULL;
CREATE INDEX idx_products_department_name ON products(department_name) WHERE department_name IS NOT NULL;
CREATE INDEX idx_products_settlement_name ON products(settlement_name) WHERE settlement_name IS NOT NULL;

-- Create a composite index for location name hierarchy searches
CREATE INDEX idx_products_location_name_hierarchy ON products(province_name, department_name, settlement_name) WHERE province_name IS NOT NULL;