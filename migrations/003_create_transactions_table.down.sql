DROP TRIGGER IF EXISTS update_product_inquiries_updated_at ON product_inquiries;
DROP TRIGGER IF EXISTS update_transactions_updated_at ON transactions;

DROP TABLE IF EXISTS product_views;
DROP TABLE IF EXISTS product_inquiries;
DROP TABLE IF EXISTS user_follows;
DROP TABLE IF EXISTS user_favorites;
DROP TABLE IF EXISTS transactions;