DROP TRIGGER IF EXISTS update_supplies_details_updated_at ON supplies_details;
DROP TRIGGER IF EXISTS update_livestock_details_updated_at ON livestock_details;
DROP TRIGGER IF EXISTS update_transport_details_updated_at ON transport_details;
DROP TRIGGER IF EXISTS update_products_updated_at ON products;

DROP TABLE IF EXISTS product_images;
DROP TABLE IF EXISTS supplies_details;
DROP TABLE IF EXISTS livestock_details;
DROP TABLE IF EXISTS transport_details;
DROP TABLE IF EXISTS products;