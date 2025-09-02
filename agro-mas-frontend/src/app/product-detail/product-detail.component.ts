import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { ProductService } from '../core/services/product.service';
import { Product, ProductImage } from '../core/models/product.model';
import { AuthService } from '../core/services/auth.service';

@Component({
  selector: 'app-product-detail',
  templateUrl: './product-detail.component.html',
  styleUrls: ['./product-detail.component.scss']
})
export class ProductDetailComponent implements OnInit {
  product: Product | null = null;
  images: ProductImage[] = [];
  currentImageIndex = 0;
  isLoading = true;
  isOwner = false;
  locationName = 'Cargando ubicación...';

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private productService: ProductService,
    public authService: AuthService
  ) { }

  ngOnInit(): void {
    const productId = this.route.snapshot.paramMap.get('id');
    if (productId) {
      this.loadProduct(productId);
    }
  }

  loadProduct(id: string): void {
    this.isLoading = true;
    this.productService.getProduct(id).subscribe({
      next: (product) => {
        if (product) {
          this.product = product;
          this.checkOwnership();
          // Use images from product instead of separate call
          this.images = (product.images || []).sort((a, b) => a.display_order - b.display_order);
          // Set location name directly
          this.locationName = this.getLocationName(product);
          this.isLoading = false;
        }
      },
      error: (error) => {
        console.error('Error loading product:', error);
        this.isLoading = false;
      }
    });
  }

  checkOwnership(): void {
    const currentUser = this.authService.currentUser;
    this.isOwner = !!(currentUser && this.product && currentUser.id === this.product.user_id);
  }

  get isAuthenticated(): boolean {
    return this.authService.isAuthenticated;
  }


  nextImage(): void {
    if (this.images.length > 1) {
      this.currentImageIndex = (this.currentImageIndex + 1) % this.images.length;
    }
  }

  prevImage(): void {
    if (this.images.length > 1) {
      this.currentImageIndex = this.currentImageIndex === 0 ? this.images.length - 1 : this.currentImageIndex - 1;
    }
  }

  getCurrentImage(): string {
    if (this.images.length > 0) {
      return this.images[this.currentImageIndex]?.image_url || this.getPlaceholderImage();
    }
    return this.getPlaceholderImage();
  }

  getPlaceholderImage(): string {
    const category = this.product?.category || 'supplies';
    const placeholders: { [key: string]: string } = {
      'transport': '/assets/images/placeholder-transport.svg',
      'livestock': '/assets/images/placeholder-livestock.svg',
      'supplies': '/assets/images/placeholder-supplies.svg'
    };
    return placeholders[category] || placeholders['supplies'];
  }

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

  contactSeller(): void {
    // Check if user is authenticated
    if (!this.authService.isAuthenticated) {
      // Redirect to login
      this.router.navigate(['/login']);
      return;
    }
    
    // TODO: Implementar funcionalidad de contacto
    console.log('Contact seller functionality to be implemented');
  }

  goToLogin(): void {
    this.router.navigate(['/login']);
  }

  editProduct(): void {
    if (this.product) {
      this.router.navigate(['/product', this.product.id], { queryParams: { edit: true } });
    }
  }

  goBack(): void {
    this.router.navigate(['/marketplace']);
  }

  getCategoryName(category: string): string {
    const categoryNames: { [key: string]: string } = {
      'transport': 'Transporte',
      'livestock': 'Ganado',
      'supplies': 'Suministros'
    };
    return categoryNames[category] || category;
  }

  getLocationName(product: Product): string {
    // Use stored location names if available
    if (product.settlement_name && product.settlement_name !== '') {
      return product.settlement_name;
    }
    
    if (product.department_name && product.department_name !== '') {
      return product.department_name;
    }
    
    if (product.province_name && product.province_name !== '') {
      return product.province_name;
    }
    
    // If we have city, use it
    if (product.city && product.city !== '') {
      return product.city;
    }
    
    // Fall back to friendly message
    return 'Argentina';
  }
}