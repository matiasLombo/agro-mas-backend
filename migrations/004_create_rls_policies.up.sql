-- Enable Row Level Security on all tables
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE products ENABLE ROW LEVEL SECURITY;
ALTER TABLE transport_details ENABLE ROW LEVEL SECURITY;
ALTER TABLE livestock_details ENABLE ROW LEVEL SECURITY;
ALTER TABLE supplies_details ENABLE ROW LEVEL SECURITY;
ALTER TABLE product_images ENABLE ROW LEVEL SECURITY;
ALTER TABLE transactions ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_favorites ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_follows ENABLE ROW LEVEL SECURITY;
ALTER TABLE product_inquiries ENABLE ROW LEVEL SECURITY;
ALTER TABLE product_views ENABLE ROW LEVEL SECURITY;

-- Users table policies
CREATE POLICY "Users can view their own profile" ON users
    FOR SELECT USING (auth.uid() = id);

CREATE POLICY "Users can update their own profile" ON users
    FOR UPDATE USING (auth.uid() = id);

CREATE POLICY "Public read access to active verified sellers" ON users
    FOR SELECT USING (role = 'seller' AND is_active = true AND is_verified = true);

CREATE POLICY "Admins can manage all users" ON users
    FOR ALL USING (
        EXISTS (
            SELECT 1 FROM users 
            WHERE id = auth.uid() AND role IN ('admin', 'moderator')
        )
    );

-- Products table policies
CREATE POLICY "Public read access to active products" ON products
    FOR SELECT USING (is_active = true AND published_at IS NOT NULL);

CREATE POLICY "Sellers can manage their own products" ON products
    FOR ALL USING (auth.uid() = user_id);

CREATE POLICY "Admins can manage all products" ON products
    FOR ALL USING (
        EXISTS (
            SELECT 1 FROM users 
            WHERE id = auth.uid() AND role IN ('admin', 'moderator')
        )
    );

-- Product detail tables policies (inherit from products)
CREATE POLICY "Public read access to transport details" ON transport_details
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM products 
            WHERE id = product_id AND is_active = true AND published_at IS NOT NULL
        )
    );

CREATE POLICY "Sellers can manage their transport details" ON transport_details
    FOR ALL USING (
        EXISTS (
            SELECT 1 FROM products 
            WHERE id = product_id AND user_id = auth.uid()
        )
    );

CREATE POLICY "Public read access to livestock details" ON livestock_details
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM products 
            WHERE id = product_id AND is_active = true AND published_at IS NOT NULL
        )
    );

CREATE POLICY "Sellers can manage their livestock details" ON livestock_details
    FOR ALL USING (
        EXISTS (
            SELECT 1 FROM products 
            WHERE id = product_id AND user_id = auth.uid()
        )
    );

CREATE POLICY "Public read access to supplies details" ON supplies_details
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM products 
            WHERE id = product_id AND is_active = true AND published_at IS NOT NULL
        )
    );

CREATE POLICY "Sellers can manage their supplies details" ON supplies_details
    FOR ALL USING (
        EXISTS (
            SELECT 1 FROM products 
            WHERE id = product_id AND user_id = auth.uid()
        )
    );

-- Product images policies
CREATE POLICY "Public read access to product images" ON product_images
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM products 
            WHERE id = product_id AND is_active = true AND published_at IS NOT NULL
        )
    );

CREATE POLICY "Sellers can manage their product images" ON product_images
    FOR ALL USING (
        EXISTS (
            SELECT 1 FROM products 
            WHERE id = product_id AND user_id = auth.uid()
        )
    );

-- Transactions table policies
CREATE POLICY "Users can view their own transactions" ON transactions
    FOR SELECT USING (auth.uid() = buyer_id OR auth.uid() = seller_id);

CREATE POLICY "Users can create transactions for products" ON transactions
    FOR INSERT WITH CHECK (auth.uid() = buyer_id);

CREATE POLICY "Transaction participants can update" ON transactions
    FOR UPDATE USING (
        auth.uid() = buyer_id OR auth.uid() = seller_id OR
        EXISTS (
            SELECT 1 FROM users 
            WHERE id = auth.uid() AND role IN ('admin', 'moderator')
        )
    );

-- User favorites policies
CREATE POLICY "Users can manage their own favorites" ON user_favorites
    FOR ALL USING (auth.uid() = user_id);

-- User follows policies
CREATE POLICY "Users can manage their own follows" ON user_follows
    FOR ALL USING (auth.uid() = follower_id);

CREATE POLICY "Public read access to follow counts" ON user_follows
    FOR SELECT USING (true);

-- Product inquiries policies
CREATE POLICY "Users can view inquiries they're involved in" ON product_inquiries
    FOR SELECT USING (auth.uid() = buyer_id OR auth.uid() = seller_id);

CREATE POLICY "Users can create inquiries" ON product_inquiries
    FOR INSERT WITH CHECK (auth.uid() = buyer_id);

CREATE POLICY "Sellers can respond to inquiries" ON product_inquiries
    FOR UPDATE USING (auth.uid() = seller_id);

-- Product views policies (limited write access for analytics)
CREATE POLICY "Users can create product views" ON product_views
    FOR INSERT WITH CHECK (true); -- Allow anonymous and authenticated views

CREATE POLICY "Users can view their own viewing history" ON product_views
    FOR SELECT USING (auth.uid() = viewer_id);

-- Auth function is now created in 000_create_auth_schema.up.sql