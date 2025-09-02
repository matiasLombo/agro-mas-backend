import { Component, OnInit, OnDestroy } from '@angular/core';
import { Router } from '@angular/router';
import { Subject, takeUntil, finalize } from 'rxjs';
import { ProductService } from '../core/services/product.service';
import { AuthService } from '../core/services/auth.service';
import { Product, ProductSearchResponse } from '@core/models/product.model';

@Component({
  selector: 'app-marketplace',
  templateUrl: './marketplace.component.html',
  styleUrls: ['./marketplace.component.scss']
})
export class MarketplaceComponent implements OnInit, OnDestroy {
  private destroy$ = new Subject<void>();

  // Data properties
  products: Product[] = [];
  isLoading = false;
  hasError = false;
  errorMessage = '';
  totalProducts = 0;

  // UI state
  showNotification = false;
  notificationMessage = '';
  notificationType: 'success' | 'info' | 'warning' | 'error' = 'info';

  // Search and filters
  searchQuery = '';
  selectedCategory = '';
  selectedLocation = '';

  constructor(
    private productService: ProductService,
    private authService: AuthService,
    private router: Router
  ) { }

  ngOnInit() {
    this.loadFeaturedProducts();
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }

  get isAuthenticated(): boolean {
    return this.authService.isAuthenticated;
  }

  get currentUser() {
    return this.authService.currentUser;
  }

  /**
   * Load featured products for the marketplace homepage
   */
  loadFeaturedProducts() {
    this.isLoading = true;
    this.hasError = false;

    this.productService.getFeaturedProducts()
      .pipe(
        takeUntil(this.destroy$),
        finalize(() => this.isLoading = false)
      )
      .subscribe({
        next: (response: ProductSearchResponse) => {
          // Filter out user's own products if logged in
          if (this.isAuthenticated && this.currentUser?.id) {
            this.products = response.products.filter(product => product.user_id !== this.currentUser!.id);
          } else {
            this.products = response.products;
          }
          this.totalProducts = response.total_count;
          this.hasError = false;
        },
        error: (error) => {
          console.error('Error loading products:', error);
          this.hasError = true;
          this.errorMessage = 'Error al cargar los productos. Por favor, intenta de nuevo.';
          this.showToast('Error al cargar productos', 'error');
          // Fallback to show some placeholder data
          this.loadPlaceholderProducts();
        }
      });
  }

  /**
   * Load products by category
   */
  loadProductsByCategory(category: string) {
    this.isLoading = true;
    this.selectedCategory = category;

    this.productService.getProductsByCategory(category)
      .pipe(
        takeUntil(this.destroy$),
        finalize(() => this.isLoading = false)
      )
      .subscribe({
        next: (response: ProductSearchResponse) => {
          this.products = response.products;
          this.totalProducts = response.total_count;
          this.hasError = false;
        },
        error: (error) => {
          console.error('Error loading products by category:', error);
          this.showToast('Error al cargar productos por categoría', 'error');
        }
      });
  }

  /**
   * Search products
   */
  searchProducts() {
    if (!this.searchQuery.trim()) {
      this.loadFeaturedProducts();
      return;
    }

    this.isLoading = true;

    this.productService.searchProducts({
      query: this.searchQuery,
      category: this.selectedCategory || undefined,
      page: 1,
      page_size: 20
    })
      .pipe(
        takeUntil(this.destroy$),
        finalize(() => this.isLoading = false)
      )
      .subscribe({
        next: (response: ProductSearchResponse) => {
          this.products = response.products;
          this.totalProducts = response.total_count;
          this.hasError = false;
        },
        error: (error) => {
          console.error('Error searching products:', error);
          this.showToast('Error en la búsqueda', 'error');
        }
      });
  }

  /**
   * Handle product interactions (requires login for non-authenticated users)
   */
  requireLogin(action: string) {
    if (this.isAuthenticated) {
      this.showToast(`Funcionalidad "${action}" en desarrollo. ¡Pronto estará disponible!`, 'info');
    } else {
      this.showToast(`Para ${action.toLowerCase()} necesitas iniciar sesión`, 'warning');
      setTimeout(() => {
        window.location.href = '/auth';
      }, 2000);
    }
  }


  /**
   * Format price for display
   */
  formatPrice(product: Product): string {
    const formatter = new Intl.NumberFormat('es-AR', {
      style: 'currency',
      currency: product.currency || 'ARS',
      minimumFractionDigits: 0
    });

    let priceText = formatter.format(product.price);
    if (product.unit) {
      priceText += ` / ${product.unit}`;
    }
    return priceText;
  }


  /**
   * Show notification toast
   */
  private showToast(message: string, type: 'success' | 'info' | 'warning' | 'error') {
    this.notificationMessage = message;
    this.notificationType = type;
    this.showNotification = true;

    setTimeout(() => {
      this.showNotification = false;
    }, 4000);
  }

  /**
   * Close notification
   */
  closeNotification() {
    this.showNotification = false;
  }

  /**
   * Get category name in Spanish
   */
  getCategoryName(category: string): string {
    const categoryNames: { [key: string]: string } = {
      'transport': 'Transporte',
      'livestock': 'Ganado',
      'supplies': 'Suministros'
    };
    return categoryNames[category] || category;
  }

  /**
   * Get primary image URL for a product
   */
  getPrimaryImage(product: Product): string {
    // First try to get the primary image
    const primaryImage = product.images?.find(img => img.is_primary);
    if (primaryImage) {
      return primaryImage.image_url;
    }

    // If no primary image, get the first available image
    const firstImage = product.images?.find(img => img.image_url);
    if (firstImage) {
      return firstImage.image_url;
    }

    // Return placeholder image based on category
    return this.getPlaceholderImage(product.category);
  }

  /**
   * Get placeholder image based on category
   */
  getPlaceholderImage(category: string): string {
    const placeholders: { [key: string]: string } = {
      'transport': 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" width="400" height="300" viewBox="0 0 400 300"><rect width="400" height="300" fill="%23f8fafc"/><g fill="%234CAF50"><text x="50%" y="50%" text-anchor="middle" font-family="Arial,sans-serif" font-size="64" dy=".3em">🚛</text></g></svg>',
      'livestock': 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" width="400" height="300" viewBox="0 0 400 300"><rect width="400" height="300" fill="%23f8fafc"/><g fill="%234CAF50"><text x="50%" y="50%" text-anchor="middle" font-family="Arial,sans-serif" font-size="64" dy=".3em">🐄</text></g></svg>',
      'supplies': 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" width="400" height="300" viewBox="0 0 400 300"><rect width="400" height="300" fill="%23f8fafc"/><g fill="%234CAF50"><text x="50%" y="50%" text-anchor="middle" font-family="Arial,sans-serif" font-size="64" dy=".3em">🌾</text></g></svg>'
    };

    return placeholders[category] || placeholders['supplies'];
  }

  /**
   * Handle image load errors
   */
  onImageError(event: any): void {
    const img = event.target;
    const productCard = img.closest('.product-card');
    if (productCard) {
      const categoryElement = productCard.querySelector('.product-category');
      const category = categoryElement?.textContent?.toLowerCase();

      // Map Spanish category names back to English for placeholder
      const categoryMap: { [key: string]: string } = {
        'transporte': 'transport',
        'ganado': 'livestock',
        'suministros': 'supplies'
      };

      const categoryKey = categoryMap[category || ''] || 'supplies';
      img.src = this.getPlaceholderImage(categoryKey);
    }
  }

  /**
   * Navigate to product details
   */
  viewProduct(productId: string): void {
    this.router.navigate(['/product-detail', productId]);
  }

  /**
   * Load placeholder products as fallback
   */
  private loadPlaceholderProducts() {
    this.products = [
      {
        id: 'placeholder-1',
        user_id: 'user-1',
        title: 'Semillas de Maíz Premium',
        description: 'Semillas híbridas de alta calidad, resistentes a sequía y plagas.',
        category: 'semillas',
        price: 25000,
        price_type: 'fixed',
        currency: 'ARS',
        unit: 'kg',
        is_active: true,
        is_featured: true,
        pickup_available: true,
        delivery_available: false,
        seller_name: 'Vendedor Demo',
        views_count: 24,
        favorites_count: 12,
        inquiries_count: 5,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        tags: ['semillas', 'maíz', 'premium']
      }
    ];
    this.totalProducts = 1;
  }
}