-- Rollback published_at fix (not recommended - this will hide existing products)
-- This would set published_at back to NULL for products that were fixed by the up migration

-- We can't reliably rollback this change without tracking which products were affected
-- For safety, we'll leave this empty
-- If you really need to rollback, you would need to manually identify which products to update