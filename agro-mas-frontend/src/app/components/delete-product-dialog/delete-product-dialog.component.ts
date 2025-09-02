import { Component, Inject } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { Product } from '../../core/models/product.model';

export interface DeleteProductDialogData {
  product: Product;
  action: 'delete' | 'toggle';
}

@Component({
  selector: 'app-delete-product-dialog',
  templateUrl: './delete-product-dialog.component.html',
  styleUrls: ['./delete-product-dialog.component.scss']
})
export class DeleteProductDialogComponent {
  isDelete: boolean;
  isToggle: boolean;
  isDeactivating: boolean;
  
  constructor(
    public dialogRef: MatDialogRef<DeleteProductDialogComponent>,
    @Inject(MAT_DIALOG_DATA) public data: DeleteProductDialogData
  ) {
    this.isDelete = data.action === 'delete';
    this.isToggle = data.action === 'toggle';
    this.isDeactivating = this.isToggle && data.product.is_active;
  }

  onCancel(): void {
    this.dialogRef.close(false);
  }

  onConfirm(): void {
    this.dialogRef.close(true);
  }

  get dialogTitle(): string {
    if (this.isDelete) {
      return '🗑️ Eliminar Producto';
    }
    return this.isDeactivating ? '⏸️ Desactivar Producto' : '▶️ Activar Producto';
  }

  get dialogMessage(): string {
    const productTitle = this.data.product.title;
    
    if (this.isDelete) {
      return `¿Estás seguro de que quieres eliminar permanentemente "${productTitle}"?`;
    }
    
    if (this.isDeactivating) {
      return `¿Deseas desactivar "${productTitle}"? Podrás reactivarlo más tarde.`;
    }
    
    return `¿Deseas reactivar "${productTitle}"? Volverá a ser visible en el marketplace.`;
  }

  get warningMessage(): string {
    if (this.isDelete) {
      return 'Esta acción no se puede deshacer. El producto será eliminado permanentemente.';
    }
    
    if (this.isDeactivating) {
      return 'El producto dejará de ser visible para los compradores hasta que lo reactives.';
    }
    
    return 'El producto volverá a ser visible en el marketplace para todos los compradores.';
  }

  get confirmButtonText(): string {
    if (this.isDelete) {
      return 'Eliminar Permanentemente';
    }
    return this.isDeactivating ? 'Desactivar' : 'Activar';
  }

  get confirmButtonColor(): string {
    if (this.isDelete) {
      return 'warn';
    }
    return this.isDeactivating ? 'accent' : 'primary';
  }

  get dialogIcon(): string {
    if (this.isDelete) {
      return '🗑️';
    }
    return this.isDeactivating ? '⏸️' : '▶️';
  }

  // Helper methods for template
  hasImage(): boolean {
    return !!(this.data.product.images?.length && this.data.product.images.some(img => img.image_url));
  }

  getPrimaryImage(): string {
    const primaryImage = this.data.product.images?.find(img => img.is_primary);
    if (primaryImage) {
      return primaryImage.image_url;
    }

    const firstImage = this.data.product.images?.find(img => img.image_url);
    if (firstImage) {
      return firstImage.image_url;
    }

    return '';
  }

  getPlaceholderEmoji(): string {
    const placeholders: { [key: string]: string } = {
      'transport': '🚚',
      'livestock': '🐄',
      'supplies': '🌾'
    };
    return placeholders[this.data.product.category] || placeholders['supplies'];
  }

  getCategoryName(): string {
    const categoryNames: { [key: string]: string } = {
      'transport': 'Transporte',
      'livestock': 'Ganado', 
      'supplies': 'Suministros'
    };
    return categoryNames[this.data.product.category] || this.data.product.category;
  }

  formatPrice(): string {
    const formatter = new Intl.NumberFormat('es-AR', {
      style: 'currency',
      currency: this.data.product.currency || 'ARS',
      minimumFractionDigits: 0
    });

    let priceText = formatter.format(this.data.product.price || 0);
    if (this.data.product.unit) {
      priceText += ` / ${this.data.product.unit}`;
    }
    return priceText;
  }
}