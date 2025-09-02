export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  phone?: string;
  cuit?: string;
  business_name?: string;
  business_type?: 'individual' | 'company' | 'cooperative';
  tax_category?: string;
  province?: string;
  city?: string;
  address?: string;
  location?: string;
  bio?: string;
  role: 'buyer' | 'seller' | 'admin' | 'moderator';
  verification_level: number; // 0-4
  is_active?: boolean;
  is_verified: boolean;
  rating: number;
  total_sales: number;
  total_purchases: number;
  total_reviews: number;
  created_at: string;
  updated_at?: string;
  preferences?: any;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  phone?: string;
  role: 'buyer' | 'seller';
  cuit?: string;
  business_name?: string;
  business_type?: 'individual' | 'company' | 'cooperative';
  province?: string;
  city?: string;
  address?: string;
}

export interface TokenInfo {
  access_token: string;
  refresh_token: string;
  expires_at: string;
  token_type: string;
}

export interface AuthResponse {
  message: string;
  user: User;
  token: TokenInfo;
}