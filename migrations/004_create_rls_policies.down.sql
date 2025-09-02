-- Drop all RLS policies
DROP POLICY IF EXISTS "Users can view their own viewing history" ON product_views;
DROP POLICY IF EXISTS "Users can create product views" ON product_views;
DROP POLICY IF EXISTS "Sellers can respond to inquiries" ON product_inquiries;
DROP POLICY IF EXISTS "Users can create inquiries" ON product_inquiries;
DROP POLICY IF EXISTS "Users can view inquiries they're involved in" ON product_inquiries;
DROP POLICY IF EXISTS "Public read access to follow counts" ON user_follows;
DROP POLICY IF EXISTS "Users can manage their own follows" ON user_follows;
DROP POLICY IF EXISTS "Users can manage their own favorites" ON user_favorites;
DROP POLICY IF EXISTS "Transaction participants can update" ON transactions;
DROP POLICY IF EXISTS "Users can create transactions for products" ON transactions;
DROP POLICY IF EXISTS "Users can view their own transactions" ON transactions;
DROP POLICY IF EXISTS "Sellers can manage their product images" ON product_images;
DROP POLICY IF EXISTS "Public read access to product images" ON product_images;
DROP POLICY IF EXISTS "Sellers can manage their supplies details" ON supplies_details;
DROP POLICY IF EXISTS "Public read access to supplies details" ON supplies_details;
DROP POLICY IF EXISTS "Sellers can manage their livestock details" ON livestock_details;
DROP POLICY IF EXISTS "Public read access to livestock details" ON livestock_details;
DROP POLICY IF EXISTS "Sellers can manage their transport details" ON transport_details;
DROP POLICY IF EXISTS "Public read access to transport details" ON transport_details;
DROP POLICY IF EXISTS "Admins can manage all products" ON products;
DROP POLICY IF EXISTS "Sellers can manage their own products" ON products;
DROP POLICY IF EXISTS "Public read access to active products" ON products;
DROP POLICY IF EXISTS "Admins can manage all users" ON users;
DROP POLICY IF EXISTS "Public read access to active verified sellers" ON users;
DROP POLICY IF EXISTS "Users can update their own profile" ON users;
DROP POLICY IF EXISTS "Users can view their own profile" ON users;

-- Disable RLS on all tables
ALTER TABLE product_views DISABLE ROW LEVEL SECURITY;
ALTER TABLE product_inquiries DISABLE ROW LEVEL SECURITY;
ALTER TABLE user_follows DISABLE ROW LEVEL SECURITY;
ALTER TABLE user_favorites DISABLE ROW LEVEL SECURITY;
ALTER TABLE transactions DISABLE ROW LEVEL SECURITY;
ALTER TABLE product_images DISABLE ROW LEVEL SECURITY;
ALTER TABLE supplies_details DISABLE ROW LEVEL SECURITY;
ALTER TABLE livestock_details DISABLE ROW LEVEL SECURITY;
ALTER TABLE transport_details DISABLE ROW LEVEL SECURITY;
ALTER TABLE products DISABLE ROW LEVEL SECURITY;
ALTER TABLE users DISABLE ROW LEVEL SECURITY;

-- Drop auth function
DROP FUNCTION IF EXISTS auth.uid();