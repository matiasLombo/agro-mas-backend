-- Create transactions table for marketplace operations
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id),
    buyer_id UUID NOT NULL REFERENCES users(id),
    seller_id UUID NOT NULL REFERENCES users(id),
    
    -- Transaction details
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'in_progress', 'completed', 'cancelled', 'disputed')),
    transaction_type VARCHAR(20) DEFAULT 'sale' CHECK (transaction_type IN ('sale', 'rental', 'service')),
    
    -- Pricing information
    original_price DECIMAL(12,2),
    negotiated_price DECIMAL(12,2),
    final_price DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'ARS',
    quantity INTEGER NOT NULL DEFAULT 1,
    unit VARCHAR(50),
    
    -- Payment information
    payment_method VARCHAR(50), -- cash, transfer, check, etc.
    payment_status VARCHAR(20) DEFAULT 'pending' CHECK (payment_status IN ('pending', 'partial', 'completed', 'refunded')),
    payment_date TIMESTAMP WITH TIME ZONE,
    
    -- Logistics information
    pickup_address TEXT,
    pickup_coordinates POINT,
    pickup_date TIMESTAMP WITH TIME ZONE,
    pickup_contact_name VARCHAR(255),
    pickup_contact_phone VARCHAR(20),
    
    delivery_address TEXT,
    delivery_coordinates POINT,
    delivery_date TIMESTAMP WITH TIME ZONE,
    delivery_contact_name VARCHAR(255),
    delivery_contact_phone VARCHAR(20),
    
    -- Communication tracking
    whatsapp_thread_id VARCHAR(255), -- Track WhatsApp conversation
    communication_log JSONB DEFAULT '[]', -- Store communication history
    
    -- Quality and feedback
    buyer_rating INTEGER CHECK (buyer_rating >= 1 AND buyer_rating <= 5),
    seller_rating INTEGER CHECK (seller_rating >= 1 AND seller_rating <= 5),
    buyer_review TEXT,
    seller_review TEXT,
    buyer_review_date TIMESTAMP WITH TIME ZONE,
    seller_review_date TIMESTAMP WITH TIME ZONE,
    
    -- Dispute handling
    dispute_reason TEXT,
    dispute_resolution TEXT,
    dispute_resolved_at TIMESTAMP WITH TIME ZONE,
    dispute_resolved_by UUID REFERENCES users(id),
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    cancellation_reason TEXT,
    
    -- Additional data
    notes TEXT,
    metadata JSONB DEFAULT '{}' -- Store additional transaction-specific data
);

-- Create user_favorites table for product bookmarking
CREATE TABLE user_favorites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id, product_id)
);

-- Create user_follows table for following sellers
CREATE TABLE user_follows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    follower_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(follower_id, following_id),
    CHECK (follower_id != following_id)
);

-- Create product_inquiries table for tracking buyer inquiries
CREATE TABLE product_inquiries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    buyer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    seller_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    inquiry_type VARCHAR(50) DEFAULT 'general' CHECK (inquiry_type IN ('general', 'price', 'availability', 'technical', 'logistics')),
    subject VARCHAR(255),
    message TEXT NOT NULL,
    
    -- Response tracking
    response TEXT,
    responded_at TIMESTAMP WITH TIME ZONE,
    is_responded BOOLEAN DEFAULT false,
    
    -- WhatsApp integration
    whatsapp_sent BOOLEAN DEFAULT false,
    whatsapp_message_id VARCHAR(255),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create product_views table for analytics
CREATE TABLE product_views (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    viewer_id UUID REFERENCES users(id) ON DELETE SET NULL, -- NULL for anonymous views
    ip_address INET,
    user_agent TEXT,
    referrer TEXT,
    viewed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    session_id VARCHAR(255),
    
    -- Geographic information
    viewer_country VARCHAR(2),
    viewer_region VARCHAR(100),
    viewer_city VARCHAR(100)
);

-- Create comprehensive indexes for transactions
CREATE INDEX idx_transactions_product_id ON transactions(product_id);
CREATE INDEX idx_transactions_buyer_id ON transactions(buyer_id);
CREATE INDEX idx_transactions_seller_id ON transactions(seller_id);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
CREATE INDEX idx_transactions_final_price ON transactions(final_price);
CREATE INDEX idx_transactions_pickup_date ON transactions(pickup_date);
CREATE INDEX idx_transactions_delivery_date ON transactions(delivery_date);

-- Indexes for favorites and follows
CREATE INDEX idx_user_favorites_user_id ON user_favorites(user_id);
CREATE INDEX idx_user_favorites_product_id ON user_favorites(product_id);
CREATE INDEX idx_user_follows_follower_id ON user_follows(follower_id);
CREATE INDEX idx_user_follows_following_id ON user_follows(following_id);

-- Indexes for inquiries
CREATE INDEX idx_product_inquiries_product_id ON product_inquiries(product_id);
CREATE INDEX idx_product_inquiries_buyer_id ON product_inquiries(buyer_id);
CREATE INDEX idx_product_inquiries_seller_id ON product_inquiries(seller_id);
CREATE INDEX idx_product_inquiries_created_at ON product_inquiries(created_at);
CREATE INDEX idx_product_inquiries_is_responded ON product_inquiries(is_responded);

-- Indexes for views (for analytics)
CREATE INDEX idx_product_views_product_id ON product_views(product_id);
CREATE INDEX idx_product_views_viewer_id ON product_views(viewer_id) WHERE viewer_id IS NOT NULL;
CREATE INDEX idx_product_views_viewed_at ON product_views(viewed_at);
CREATE INDEX idx_product_views_session_id ON product_views(session_id);

-- Add triggers for updated_at columns
CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON transactions
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_product_inquiries_updated_at BEFORE UPDATE ON product_inquiries
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();