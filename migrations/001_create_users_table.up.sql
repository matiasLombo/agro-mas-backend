-- Enable PostGIS extension for geospatial support
CREATE EXTENSION IF NOT EXISTS postgis;

-- Create users table with comprehensive agricultural marketplace support
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    
    -- Personal Information
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    
    -- Argentine-specific business information
    cuit VARCHAR(11) UNIQUE, -- Argentine tax ID
    business_name VARCHAR(255),
    business_type VARCHAR(50), -- "individual", "company", "cooperative"
    tax_category VARCHAR(50), -- AFIP categories
    
    -- Geographic information
    province VARCHAR(100),
    city VARCHAR(100),
    address TEXT,
    coordinates POINT, -- PostGIS point for geospatial queries
    
    -- User roles and verification
    role VARCHAR(20) DEFAULT 'buyer' CHECK (role IN ('buyer', 'seller', 'admin', 'moderator')),
    verification_level INTEGER DEFAULT 0, -- 0=none, 1=email, 2=cuit, 3=business, 4=field
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,
    
    -- Denormalized statistics for performance
    rating DECIMAL(3,2) DEFAULT 0.0,
    total_sales INTEGER DEFAULT 0,
    total_purchases INTEGER DEFAULT 0,
    total_reviews INTEGER DEFAULT 0,
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login TIMESTAMP WITH TIME ZONE,
    verification_documents JSONB, -- Store document URLs and metadata
    preferences JSONB DEFAULT '{}' -- User preferences and settings
);

-- Create indexes for optimized queries
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_cuit ON users(cuit) WHERE cuit IS NOT NULL;
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_province ON users(province);
CREATE INDEX idx_users_verification_level ON users(verification_level);
CREATE INDEX idx_users_is_active ON users(is_active) WHERE is_active = true;
CREATE INDEX idx_users_coordinates ON users USING GIST(coordinates) WHERE coordinates IS NOT NULL;
CREATE INDEX idx_users_created_at ON users(created_at);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to automatically update updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();