import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import { Product, ProductSearchRequest, ProductSearchResponse, ProductImage } from '@core/models/product.model';

@Injectable({
  providedIn: 'root'
})
export class ProductService {
  private apiUrl = `${environment.apiUrl}/products`;

  constructor(private http: HttpClient) { }

  /**
   * Search products with filters and pagination
   */
  searchProducts(request: ProductSearchRequest = {}): Observable<ProductSearchResponse> {
    let params = new HttpParams();

    if (request.query) params = params.set('query', request.query);
    if (request.category) params = params.set('category', request.category);
    if (request.subcategory) params = params.set('subcategory', request.subcategory);
    if (request.province) params = params.set('province', request.province);
    if (request.city) params = params.set('city', request.city);
    if (request.min_price !== undefined) params = params.set('min_price', request.min_price.toString());
    if (request.max_price !== undefined) params = params.set('max_price', request.max_price.toString());
    if (request.price_type) params = params.set('price_type', request.price_type);
    if (request.pickup_available !== undefined) params = params.set('pickup_available', request.pickup_available.toString());
    if (request.delivery_available !== undefined) params = params.set('delivery_available', request.delivery_available.toString());
    if (request.is_verified_seller !== undefined) params = params.set('is_verified_seller', request.is_verified_seller.toString());
    if (request.tags && request.tags.length > 0) {
      request.tags.forEach(tag => params = params.append('tags', tag));
    }
    if (request.sort_by) params = params.set('sort_by', request.sort_by);
    if (request.page) params = params.set('page', request.page.toString());
    if (request.page_size) params = params.set('page_size', request.page_size.toString());

    return this.http.get<ProductSearchResponse>(`${this.apiUrl}/search`, { params });
  }

  /**
   * Get a single product by ID
   */
  getProduct(id: string): Observable<Product> {
    return this.http.get<Product>(`${this.apiUrl}/${id}`);
  }

  /**
   * Get featured products (first page with default filters)
   */
  getFeaturedProducts(): Observable<ProductSearchResponse> {
    return this.searchProducts({
      sort_by: 'featured',
      page: 1,
      page_size: 20  // Show more products
    });
  }

  /**
   * Get products by category
   */
  getProductsByCategory(category: string, page = 1, pageSize = 20): Observable<ProductSearchResponse> {
    return this.searchProducts({
      category,
      page,
      page_size: pageSize,
      sort_by: 'date_desc'
    });
  }


  /**
   * Get a resized image URL
   */
  getResizedImageUrl(storagePath: string, width: number, height: number, quality = 80): string {
    const baseUrl = environment.apiUrl.replace('/api', '');
    return `${baseUrl}/resize/${storagePath}?w=${width}&h=${height}&q=${quality}`;
  }

  /**
   * Get current user's products
   */
  getProductsByUser(): Observable<ProductSearchResponse> {
    return this.http.get<ProductSearchResponse>(`${this.apiUrl}/my`);
  }

  /**
   * Create a new product
   */
  createProduct(product: Partial<Product>): Observable<Product> {
    return this.http.post<Product>(`${this.apiUrl}/`, product);
  }

  /**
   * Create product with images using FormData
   */
  createProductWithImages(product: Partial<Product>, images: File[]): Observable<Product> {
    const formData = new FormData();
    
    // Add product data as JSON string
    formData.append('product', JSON.stringify(product));
    
    // Add each image file
    images.forEach((image, index) => {
      formData.append('image', image, image.name);
    });
    
    return this.http.post<Product>(`${this.apiUrl}/`, formData);
  }

  /**
   * Update an existing product
   */
  updateProduct(productId: string, updates: Partial<Product>): Observable<Product> {
    return this.http.put<Product>(`${this.apiUrl}/${productId}`, updates);
  }

  /**
   * Update product with new images using FormData
   */
  updateProductWithImages(productId: string, product: Partial<Product>, images: File[]): Observable<Product> {
    const formData = new FormData();
    
    // Clean product data - remove undefined/null values
    const cleanProduct = Object.entries(product).reduce((acc, [key, value]) => {
      if (value !== undefined && value !== null) {
        acc[key] = value;
      }
      return acc;
    }, {} as any);
    
    console.log('Product data being sent:', cleanProduct);
    
    // Add product data as JSON string
    const jsonString = JSON.stringify(cleanProduct);
    console.log('JSON string being sent:', jsonString);
    console.log('JSON string length:', jsonString.length);
    
    // Check for problematic characters
    for (let i = 0; i < jsonString.length; i++) {
      const char = jsonString[i];
      if (char === '-' && i > 0) {
        const before = jsonString.substring(Math.max(0, i-10), i);
        const after = jsonString.substring(i, Math.min(jsonString.length, i+10));
        console.log(`Found - at position ${i}: "${before}${after}"`);
      }
    }
    
    formData.append('product', jsonString);
    
    // Add each image file
    images.forEach((image, index) => {
      formData.append('image', image, image.name);
    });
    
    return this.http.put<Product>(`${this.apiUrl}/${productId}`, formData);
  }

  /**
   * Delete a product
   */
  deleteProduct(productId: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${productId}`);
  }

}