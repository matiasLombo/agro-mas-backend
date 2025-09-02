export interface Product {
    id: string;
    user_id: string;
    title: string;
    description: string;
    category: string;
    subcategory?: string;
    price: number;
    price_type: string;
    currency: string;
    unit?: string;
    quantity?: number;
    available_from?: string;
    available_until?: string;
    is_active: boolean;
    is_featured: boolean;
    province?: string;
    department?: string;
    settlement?: string;
    city?: string;
    province_name?: string;
    department_name?: string;
    settlement_name?: string;
    location_coordinates?: {
        lng: number;
        lat: number;
    };
    pickup_available: boolean;
    delivery_available: boolean;
    delivery_radius?: number;
    seller_name: string;
    seller_phone?: string;
    seller_rating?: number;
    seller_verification_level?: string;
    views_count: number;
    favorites_count: number;
    inquiries_count: number;
    search_keywords?: string;
    created_at: string;
    updated_at: string;
    published_at?: string;
    expires_at?: string;
    metadata?: any;
    tags: string[];
    images?: ProductImage[];
}

export interface ProductImage {
    id: string;
    product_id: string;
    image_url: string;
    cloud_storage_path?: string;
    alt_text?: string;
    is_primary: boolean;
    display_order: number;
    file_size?: number;
    mime_type?: string;
    uploaded_at: string;
}

export interface ProductSearchRequest {
    query?: string;
    category?: string;
    subcategory?: string;
    province?: string;
    city?: string;
    min_price?: number;
    max_price?: number;
    price_type?: string;
    pickup_available?: boolean;
    delivery_available?: boolean;
    is_verified_seller?: boolean;
    tags?: string[];
    sort_by?: string;
    page?: number;
    page_size?: number;
}

export interface ProductSearchResponse {
    products: Product[];
    total_count: number;
    page: number;
    page_size: number;
    total_pages: number;
}
