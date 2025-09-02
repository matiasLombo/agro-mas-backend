-- Create products table for agricultural marketplace
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Basic product information
    title VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL CHECK (category IN ('transport', 'livestock', 'supplies')),
    subcategory VARCHAR(100),
    
    -- Pricing information
    price DECIMAL(12,2),
    price_type VARCHAR(20) DEFAULT 'fixed' CHECK (price_type IN ('fixed', 'negotiable', 'per_unit', 'quote')),
    currency VARCHAR(3) DEFAULT 'ARS',
    unit VARCHAR(50), -- kg, ton, head, liter, etc.
    
    -- Availability
    quantity INTEGER,
    available_from DATE,
    available_until DATE,
    is_active BOOLEAN DEFAULT true,
    is_featured BOOLEAN DEFAULT false,
    
    -- Location information
    province VARCHAR(100),
    city VARCHAR(100),
    location_coordinates POINT,
    pickup_available BOOLEAN DEFAULT true,
    delivery_available BOOLEAN DEFAULT false,
    delivery_radius INTEGER, -- in kilometers
    
    -- Denormalized seller information for performance
    seller_name VARCHAR(255),
    seller_phone VARCHAR(20),
    seller_rating DECIMAL(3,2),
    seller_verification_level INTEGER,
    
    -- Engagement metrics
    views_count INTEGER DEFAULT 0,
    favorites_count INTEGER DEFAULT 0,
    inquiries_count INTEGER DEFAULT 0,
    
    -- Search optimization
    search_keywords TEXT, -- Generated keywords for search
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    published_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    
    -- Additional data storage
    metadata JSONB DEFAULT '{}', -- Store category-specific data
    tags TEXT[] -- Array of tags for flexible categorization
);

-- Create transport_details table for transport-specific information
CREATE TABLE transport_details (
    product_id UUID PRIMARY KEY REFERENCES products(id) ON DELETE CASCADE,
    vehicle_type VARCHAR(50), -- truck, trailer, pickup, etc.
    capacity_tons DECIMAL(8,2),
    capacity_cubic_meters DECIMAL(8,2),
    price_per_km DECIMAL(8,2),
    has_refrigeration BOOLEAN DEFAULT false,
    has_livestock_equipment BOOLEAN DEFAULT false,
    
    -- Service areas
    service_provinces TEXT[], -- Array of provinces where service is available
    min_distance_km INTEGER,
    max_distance_km INTEGER,
    
    -- Vehicle specifications
    license_plate VARCHAR(20),
    license_expiry DATE,
    insurance_expiry DATE,
    vehicle_year INTEGER,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create livestock_details table for livestock-specific information
CREATE TABLE livestock_details (
    product_id UUID PRIMARY KEY REFERENCES products(id) ON DELETE CASCADE,
    animal_type VARCHAR(50), -- cattle, sheep, pigs, horses, etc.
    breed VARCHAR(100),
    age_months INTEGER,
    weight_kg DECIMAL(8,2),
    gender VARCHAR(10) CHECK (gender IN ('male', 'female', 'mixed')),
    
    -- Health and certifications
    health_certificates TEXT[], -- Array of certificate URLs
    vaccinations JSONB, -- Store vaccination records
    last_veterinary_check DATE,
    is_organic BOOLEAN DEFAULT false,
    
    -- Breeding information
    is_pregnant BOOLEAN,
    breeding_history JSONB,
    genetic_information TEXT,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create supplies_details table for agricultural supplies
CREATE TABLE supplies_details (
    product_id UUID PRIMARY KEY REFERENCES products(id) ON DELETE CASCADE,
    supply_type VARCHAR(50), -- fertilizer, pesticide, seeds, equipment, etc.
    brand VARCHAR(100),
    model VARCHAR(100),
    
    -- Chemical/biological information
    active_ingredients TEXT[],
    concentration VARCHAR(50),
    expiry_date DATE,
    batch_number VARCHAR(50),
    
    -- Regulatory information
    registration_number VARCHAR(100), -- SENASA registration
    required_licenses TEXT[], -- Required licenses to purchase
    safety_data_sheet_url TEXT,
    
    -- Storage and handling
    storage_requirements TEXT,
    handling_instructions TEXT,
    disposal_instructions TEXT,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create product_images table for image management
CREATE TABLE product_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    cloud_storage_path TEXT NOT NULL,
    alt_text VARCHAR(255),
    is_primary BOOLEAN DEFAULT false,
    display_order INTEGER DEFAULT 0,
    file_size INTEGER, -- in bytes
    mime_type VARCHAR(100),
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create comprehensive indexes for products
CREATE INDEX idx_products_user_id ON products(user_id);
CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_category_subcategory ON products(category, subcategory);
CREATE INDEX idx_products_province ON products(province);
CREATE INDEX idx_products_is_active ON products(is_active) WHERE is_active = true;
CREATE INDEX idx_products_published_at ON products(published_at) WHERE published_at IS NOT NULL;
CREATE INDEX idx_products_price ON products(price) WHERE price IS NOT NULL;
CREATE INDEX idx_products_available_from ON products(available_from);
CREATE INDEX idx_products_coordinates ON products USING GIST(location_coordinates) WHERE location_coordinates IS NOT NULL;
CREATE INDEX idx_products_search ON products USING GIN(to_tsvector('spanish', title || ' ' || COALESCE(description, '') || ' ' || COALESCE(search_keywords, '')));

-- Indexes for detail tables
CREATE INDEX idx_transport_details_vehicle_type ON transport_details(vehicle_type);
CREATE INDEX idx_transport_details_capacity ON transport_details(capacity_tons);
CREATE INDEX idx_livestock_details_animal_type ON livestock_details(animal_type);
CREATE INDEX idx_livestock_details_breed ON livestock_details(breed);
CREATE INDEX idx_supplies_details_supply_type ON supplies_details(supply_type);
CREATE INDEX idx_supplies_details_brand ON supplies_details(brand);

-- Indexes for product images
CREATE INDEX idx_product_images_product_id ON product_images(product_id);
CREATE INDEX idx_product_images_is_primary ON product_images(product_id, is_primary) WHERE is_primary = true;

-- Add triggers for updated_at columns
CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_transport_details_updated_at BEFORE UPDATE ON transport_details
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_livestock_details_updated_at BEFORE UPDATE ON livestock_details
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_supplies_details_updated_at BEFORE UPDATE ON supplies_details
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();