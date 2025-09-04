-- Add department and settlement columns to products table
ALTER TABLE products ADD COLUMN department VARCHAR(100);
ALTER TABLE products ADD COLUMN settlement VARCHAR(100);

-- Create indexes for the new location columns
CREATE INDEX idx_products_department ON products(department) WHERE department IS NOT NULL;
CREATE INDEX idx_products_settlement ON products(settlement) WHERE settlement IS NOT NULL;
CREATE INDEX idx_products_location_hierarchy ON products(province, department, settlement) WHERE province IS NOT NULL;