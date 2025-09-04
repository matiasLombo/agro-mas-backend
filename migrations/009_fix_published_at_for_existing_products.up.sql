-- Fix published_at for existing active products
-- This migration ensures all active products have published_at set so they appear in searches

UPDATE products 
SET published_at = created_at 
WHERE is_active = true 
AND published_at IS NULL;