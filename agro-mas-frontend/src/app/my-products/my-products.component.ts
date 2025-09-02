import { Component, OnInit, OnDestroy } from '@angular/core';
import { Router } from '@angular/router';
import { Subject, takeUntil } from 'rxjs';
import { ProductService } from '../core/services/product.service';
import { AuthService } from '../core/services/auth.service';
import { Product } from '../core/models/product.model';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { DeleteProductDialogComponent } from '../components/delete-product-dialog/delete-product-dialog.component';

@Component({
  selector: 'app-my-products',
  templateUrl: './my-products.component.html',
  styleUrls: ['./my-products.component.scss']
})
export class MyProductsComponent implements OnInit, OnDestroy {
  private destroy$ = new Subject<void>();
  
  products: Product[] = [];
  isLoading = false;
  hasError = false;
  errorMessage = '';

  constructor(
    private productService: ProductService,
    private authService: AuthService,
    private router: Router,
    private dialog: MatDialog,
    private snackBar: MatSnackBar
  ) {}

  ngOnInit(): void {
    this.loadMyProducts();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  loadMyProducts(): void {
    const currentUser = this.authService.currentUser;
    if (!currentUser) {
      this.router.navigate(['/login']);
      return;
    }

    this.isLoading = true;
    this.hasError = false;

    this.productService.getProductsByUser()
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (response) => {
          this.products = response.products || [];
          this.isLoading = false;
        },
        error: (error) => {
          console.error('Error loading user products:', error);
          this.hasError = true;
          this.errorMessage = 'Error al cargar tus productos';
          this.isLoading = false;
          this.snackBar.open('Error al cargar productos', 'Cerrar', { duration: 3000 });
        }
      });
  }

  editProduct(productId: string): void {
    this.router.navigate(['/product', productId], { queryParams: { edit: true } });
  }

  viewProduct(productId: string): void {
    this.router.navigate(['/product-detail', productId]);
  }

  toggleProductStatus(product: Product): void {
    const dialogRef = this.dialog.open(DeleteProductDialogComponent, {
      width: '540px',
      maxWidth: '95vw',
      data: {
        product: product,
        action: 'toggle'
      },
      disableClose: false,
      panelClass: ['custom-dialog-container'],
      backdropClass: 'custom-backdrop'
    });

    dialogRef.afterClosed().subscribe(result => {
      if (result === true) {
        const newStatus = !product.is_active;
        this.productService.updateProduct(product.id, { is_active: newStatus })
          .subscribe({
            next: () => {
              product.is_active = newStatus;
              const message = newStatus ? '✅ Producto activado exitosamente' : '⏸️ Producto desactivado exitosamente';
              this.snackBar.open(message, 'Cerrar', { 
                duration: 4000,
                panelClass: newStatus ? ['success-snackbar'] : ['info-snackbar']
              });
            },
            error: (error) => {
              console.error('Error updating product status:', error);
              this.snackBar.open('❌ Error al actualizar el estado del producto', 'Cerrar', { 
                duration: 4000,
                panelClass: ['error-snackbar']
              });
            }
          });
      }
    });
  }

  deleteProduct(product: Product): void {
    const dialogRef = this.dialog.open(DeleteProductDialogComponent, {
      width: '540px',
      maxWidth: '95vw',
      data: {
        product: product,
        action: 'delete'
      },
      disableClose: false,
      panelClass: ['custom-dialog-container'],
      backdropClass: 'custom-backdrop'
    });

    dialogRef.afterClosed().subscribe(result => {
      if (result === true) {
        this.productService.deleteProduct(product.id)
          .subscribe({
            next: () => {
              this.products = this.products.filter(p => p.id !== product.id);
              this.snackBar.open('🗑️ Producto eliminado permanentemente', 'Cerrar', { 
                duration: 4000,
                panelClass: ['success-snackbar']
              });
            },
            error: (error) => {
              console.error('Error deleting product:', error);
              this.snackBar.open('❌ Error al eliminar el producto', 'Cerrar', { 
                duration: 4000,
                panelClass: ['error-snackbar']
              });
            }
          });
      }
    });
  }

  createNewProduct(): void {
    this.router.navigate(['/product']);
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

  getPrimaryImage(product: Product): string {
    const primaryImage = product.images?.find(img => img.is_primary);
    if (primaryImage) {
      return primaryImage.image_url;
    }

    const firstImage = product.images?.find(img => img.image_url);
    if (firstImage) {
      return firstImage.image_url;
    }

    return this.getPlaceholderImage(product.category);
  }

  getPlaceholderImage(category: string): string {
    return '';
  }

  getPlaceholderEmoji(category: string): string {
    const placeholders: { [key: string]: string } = {
      'transport': '🚚',
      'livestock': '🐄',
      'supplies': '🌾'
    };
    return placeholders[category] || placeholders['supplies'];
  }

  hasImage(product: Product): boolean {
    return !!(product.images?.length && product.images.some(img => img.image_url));
  }

  getCategoryName(category: string): string {
    const categoryNames: { [key: string]: string } = {
      'transport': 'Transporte',
      'livestock': 'Ganado', 
      'supplies': 'Suministros'
    };
    return categoryNames[category] || category;
  }

  getStatusText(product: Product): string {
    return product.is_active ? 'Activo' : 'Inactivo';
  }

  getStatusColor(product: Product): string {
    return product.is_active ? 'primary' : 'warn';
  }

  formatDate(dateString: string): string {
    return new Date(dateString).toLocaleDateString('es-AR', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }

  // Getters for template calculations
  get activeProductsCount(): number {
    return this.products.filter(p => p.is_active).length;
  }

}