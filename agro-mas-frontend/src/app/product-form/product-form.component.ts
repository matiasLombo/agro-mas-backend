import { Component, OnInit, Input, ViewChild, ElementRef } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { Router, ActivatedRoute } from '@angular/router';
import { ProductService } from '../core/services/product.service';
import { LocationService } from '../core/services/location.service';
import { SellerService } from '../services/seller.service';
import { Product, ProductImage } from '../core/models/product.model';
import { Province, Department, Settlement } from '../core/models/location.model';
import { SellerSetupModalComponent } from '../components/seller-setup-modal/seller-setup-modal.component';
import { Observable, forkJoin } from 'rxjs';

interface FilePreview {
  file: File;
  url: string;
  type: 'image' | 'video';
  compressed?: boolean;
  originalSize: number;
  compressedSize?: number;
}

@Component({
  selector: 'app-product-form',
  templateUrl: './product-form.component.html',
  styleUrls: ['./product-form.component.scss']
})
export class ProductFormComponent implements OnInit {
  @Input() productId: string | null = null;

  productForm!: FormGroup;
  isEditMode = false;
  product: Product | null = null;
  categories = ['Ganado', 'Insumo', 'Transporte'];

  uploadedImages: ProductImage[] = [];
  uploadedVideos: any[] = [];
  filePreviews: FilePreview[] = [];
  
  isUploading: boolean = false;
  uploadProgress: number = 0;
  isDragOver: boolean = false;
  
  // File compression settings
  readonly MAX_IMAGE_SIZE = 1024 * 1024; // 1MB
  readonly MAX_VIDEO_SIZE = 10 * 1024 * 1024; // 10MB
  readonly IMAGE_QUALITY = 0.8;
  readonly MAX_IMAGE_WIDTH = 1920;
  readonly MAX_IMAGE_HEIGHT = 1080;
  
  // Make Math available in template
  Math = Math;
  
  // Template helper methods to avoid inline arrow functions
  get imageCount(): number {
    return this.filePreviews.filter(f => f.type === 'image').length;
  }
  
  get videoCount(): number {
    return this.filePreviews.filter(f => f.type === 'video').length;
  }
  
  get isImageUploadDisabled(): boolean {
    return this.imageCount >= 10;
  }
  
  get isVideoUploadDisabled(): boolean {
    return this.videoCount >= 3;
  }
  
  get hasChanges(): boolean {
    // In create mode, always allow submission if form is valid
    if (!this.isEditMode) {
      console.log('hasChanges: CREATE MODE - always true');
      return true;
    }
    
    // In edit mode, check for form changes or new images
    const formChanged = this.productForm.dirty;
    const hasNewImages = this.filePreviews.length > 0;
    
    console.log('hasChanges DEBUG:', {
      isEditMode: this.isEditMode,
      formChanged,
      hasNewImages,
      filePreviewsLength: this.filePreviews.length,
      filePreviewsContent: this.filePreviews,
      formInvalid: this.productForm.invalid,
      isUploading: this.isUploading
    });
    
    const result = formChanged || hasNewImages;
    console.log('hasChanges RESULT:', result);
    
    return result;
  }

  // Location data
  provinces: Province[] = [];
  departments: Department[] = [];
  settlements: Settlement[] = [];

  selectedProvinceId: string = '';
  selectedDepartmentId: string = '';
  selectedSettlementId: string = '';

  constructor(
    private fb: FormBuilder,
    private productService: ProductService,
    private locationService: LocationService,
    private sellerService: SellerService,
    private dialog: MatDialog,
    private snackBar: MatSnackBar,
    private router: Router,
    private route: ActivatedRoute
  ) { 
    console.log('ProductFormComponent constructor called - hasChanges getter should be available');
  }

  ngOnInit(): void {
    // Check seller profile status when component initializes
    this.checkSellerProfileAndShowModal();
    
    this.productForm = this.fb.group({
      title: ['', Validators.required],
      description: [''],
      category: ['', Validators.required],
      price: ['', [Validators.required, Validators.min(0)]],
      price_type: ['fixed', Validators.required],
      currency: ['ARS', Validators.required],
      province: ['', Validators.required],
      department: ['', Validators.required],
      settlement: ['', Validators.required],
      city: [''],
      location_coordinates: this.fb.group({
        lat: [''],
        lng: ['']
      }),
      pickup_available: [false],
      delivery_available: [false],
      user_id: [''],
      is_active: [true],
      is_featured: [false],
      seller_name: [''],
      tags: [[]],
      images: [[]],
    });

    this.loadProvinces();

    // Check for product ID from route parameter
    const routeProductId = this.route.snapshot.paramMap.get('id');
    const isEditQueryParam = this.route.snapshot.queryParams['edit'];
    
    if (routeProductId) {
      this.productId = routeProductId;
      // Solo permitir edición si el query param 'edit' está presente
      this.isEditMode = isEditQueryParam === 'true';
      this.loadProductData(this.productId);
    }
  }

  loadProductData(id: string): void {
    console.log('Loading product data for ID:', id);
    this.productService.getProduct(id).subscribe({
      next: (response: any) => {
        console.log('Product data received:', response);
        const product = response.product || response;
        console.log('Extracted product:', product);
        console.log('Department:', product.department, 'Settlement:', product.settlement, 'Province:', product.province);
        if (product) {
          this.product = product;
          // Map backend categories to frontend categories
          const categoryMapping: { [key: string]: string } = {
            'livestock': 'Ganado',
            'supplies': 'Insumo',
            'transport': 'Transporte'
          };
          
          // Update form with product data
          const formData = {
            title: product.title,
            description: product.description,
            category: categoryMapping[product.category] || product.category,
            price: product.price,
            price_type: product.price_type,
            currency: product.currency,
            province: product.province,
            department: product.department,
            settlement: product.settlement,
            city: product.city,
            pickup_available: product.pickup_available,
            delivery_available: product.delivery_available,
            seller_name: product.seller_name,
            tags: product.tags || [],
            images: product.images || []
          };
          console.log('Form data to patch:', formData);
          this.productForm.patchValue(formData);

          // Load location data if province exists
          if (product.province) {
            this.loadLocationDataForEdit(product.province, product.department, product.settlement);
          }

          // Load product images after product is set
          this.loadProductImages();

          // Only enable form if we're in edit mode, otherwise disable for viewing
          if (!this.isEditMode) {
            this.productForm.disable();
          }
        }
      },
      error: (error) => {
        console.error('Error loading product:', error);
      }
    });
  }

  loadProductImages(): void {
    // Images are now loaded from the product directly
    if (this.product?.images) {
      this.uploadedImages = [...this.product.images];
      console.log('Loaded existing images:', this.uploadedImages);
    } else {
      this.uploadedImages = [];
    }
  }

  async onFileChange(event: any, type: 'image' | 'video'): Promise<void> {
    const files: FileList = event.target.files;
    if (files.length === 0) return;

    // Validate file count
    const currentFileCount = this.filePreviews.length;
    const maxFiles = type === 'image' ? 10 : 3;
    
    if (currentFileCount + files.length > maxFiles) {
      this.showNotification(`Máximo ${maxFiles} ${type === 'image' ? 'imágenes' : 'videos'} permitidos`, 'error');
      return;
    }

    this.isUploading = true;
    this.uploadProgress = 0;

    try {
      // Process files with compression
      for (let i = 0; i < files.length; i++) {
        const file = files[i];
        const progress = ((i + 1) / files.length) * 50; // First 50% for processing
        this.uploadProgress = progress;

        const preview = await this.processFile(file, type);
        if (preview) {
          this.filePreviews.push(preview);
        }
      }

      // Upload files if in edit mode
      if (this.isEditMode && this.productId) {
        await this.uploadProcessedFiles();
      }

    } catch (error) {
      console.error('Error processing files:', error);
      this.showNotification('Error al procesar archivos', 'error');
    } finally {
      this.isUploading = false;
      this.uploadProgress = 0;
      // Clear file input
      event.target.value = '';
    }
  }

  private async processFile(file: File, type: 'image' | 'video'): Promise<FilePreview | null> {
    try {
      // Validate file type
      if (type === 'image' && !file.type.startsWith('image/')) {
        this.showNotification(`${file.name} no es una imagen válida`, 'error');
        return null;
      }
      if (type === 'video' && !file.type.startsWith('video/')) {
        this.showNotification(`${file.name} no es un video válido`, 'error');
        return null;
      }

      // Create preview URL
      const url = URL.createObjectURL(file);
      const originalSize = file.size;

      let processedFile = file;
      let compressedSize = originalSize;
      let compressed = false;

      // Compress images
      if (type === 'image' && file.size > this.MAX_IMAGE_SIZE) {
        try {
          processedFile = await this.compressImage(file);
          compressedSize = processedFile.size;
          compressed = true;
        } catch (error) {
          console.warn('Image compression failed, using original:', error);
        }
      }

      // Validate video size
      if (type === 'video' && file.size > this.MAX_VIDEO_SIZE) {
        this.showNotification(`El video ${file.name} es demasiado grande (máximo 10MB)`, 'error');
        return null;
      }

      return {
        file: processedFile,
        url,
        type,
        compressed,
        originalSize,
        compressedSize
      };

    } catch (error) {
      console.error('Error processing file:', error);
      this.showNotification(`Error al procesar ${file.name}`, 'error');
      return null;
    }
  }

  private async compressImage(file: File): Promise<File> {
    return new Promise((resolve, reject) => {
      const canvas = document.createElement('canvas');
      const ctx = canvas.getContext('2d');
      const img = new Image();

      img.onload = () => {
        try {
          // Calculate new dimensions
          let { width, height } = img;
          const ratio = Math.min(this.MAX_IMAGE_WIDTH / width, this.MAX_IMAGE_HEIGHT / height);
          
          if (ratio < 1) {
            width *= ratio;
            height *= ratio;
          }

          // Set canvas size
          canvas.width = width;
          canvas.height = height;

          // Draw and compress
          ctx?.drawImage(img, 0, 0, width, height);
          
          canvas.toBlob(
            (blob) => {
              if (blob) {
                const compressedFile = new File([blob], file.name, {
                  type: 'image/jpeg',
                  lastModified: Date.now()
                });
                resolve(compressedFile);
              } else {
                reject(new Error('Failed to compress image'));
              }
            },
            'image/jpeg',
            this.IMAGE_QUALITY
          );
        } catch (error) {
          reject(error);
        }
      };

      img.onerror = () => reject(new Error('Failed to load image'));
      img.src = URL.createObjectURL(file);
    });
  }

  private async uploadProcessedFiles(): Promise<void> {
    if (!this.productId) return;

    const filesToUpload = this.filePreviews.filter(preview => 
      preview.type === 'image' // Only upload images for now
    );

    for (let i = 0; i < filesToUpload.length; i++) {
      const preview = filesToUpload[i];
      const progress = 50 + ((i + 1) / filesToUpload.length) * 50; // Second 50% for uploading
      this.uploadProgress = progress;

      try {
        const isPrimary = this.uploadedImages.length === 0 && i === 0;
        const displayOrder = this.uploadedImages.length + i + 1;
        
        // Images are now processed during product creation/update, not separately
        console.log('Image will be processed with product creation/update');
      } catch (error) {
        console.error(`Error uploading ${preview.file.name}:`, error);
        this.showNotification(`Error al subir ${preview.file.name}`, 'error');
      }
    }
    
    // Update form
    this.productForm.get('images')?.setValue(this.uploadedImages);
    this.showNotification(`${filesToUpload.length} archivo(s) subido(s) exitosamente`, 'success');
  }

  removeFile(index: number): void {
    const preview = this.filePreviews[index];
    if (preview) {
      // Revoke object URL to free memory
      URL.revokeObjectURL(preview.url);
      this.filePreviews.splice(index, 1);
    }
  }

  getFileSizeText(bytes: number): string {
    if (bytes < 1024) return bytes + ' B';
    else if (bytes < 1048576) return Math.round(bytes / 1024) + ' KB';
    else return Math.round(bytes / 1048576) + ' MB';
  }

  private showNotification(message: string, type: 'success' | 'error' | 'warning' = 'success'): void {
    this.snackBar.open(message, 'Cerrar', {
      duration: type === 'error' ? 5000 : 3000,
      panelClass: [`snackbar-${type}`]
    });
  }

  // Drag & Drop handlers
  onDragOver(event: DragEvent): void {
    event.preventDefault();
    event.stopPropagation();
    this.isDragOver = true;
  }

  onDragLeave(event: DragEvent): void {
    event.preventDefault();
    event.stopPropagation();
    this.isDragOver = false;
  }

  onDrop(event: DragEvent): void {
    event.preventDefault();
    event.stopPropagation();
    this.isDragOver = false;

    const files = event.dataTransfer?.files;
    if (files && files.length > 0) {
      this.handleFiles(files);
    }
  }

  private handleFiles(files: FileList): void {
    const imageFiles: File[] = [];
    const videoFiles: File[] = [];

    // Separar archivos por tipo
    Array.from(files).forEach(file => {
      if (file.type.startsWith('image/')) {
        imageFiles.push(file);
      } else if (file.type.startsWith('video/')) {
        videoFiles.push(file);
      }
    });

    // Procesar imágenes
    if (imageFiles.length > 0) {
      const event = { target: { files: imageFiles } } as any;
      this.onFileChange(event, 'image');
    }

    // Procesar videos
    if (videoFiles.length > 0) {
      const event = { target: { files: videoFiles } } as any;
      this.onFileChange(event, 'video');
    }
  }

  onSubmit(): void {
    if (this.productForm.valid && !this.isUploading) {
      // Get form values INCLUDING disabled fields
      const formValue = this.productForm.getRawValue();
      console.log('Form values:', formValue);
      
      // Map frontend categories to backend categories
      const categoryMapping: { [key: string]: string } = {
        'Ganado': 'livestock',
        'Insumo': 'supplies',
        'Transporte': 'transport'
      };

      const mappedCategory = categoryMapping[formValue.category] || formValue.category.toLowerCase();
      
      const productData: any = {
        title: formValue.title,
        description: formValue.description,
        category: mappedCategory,
        price: parseFloat(formValue.price),
        price_type: formValue.price_type,
        currency: formValue.currency || 'ARS',
        province: formValue.province,
        department: formValue.department,
        settlement: formValue.settlement,
        city: formValue.city || this.getSettlementName(formValue.settlement),
        // Add location names for display
        province_name: this.getProvinceName(formValue.province),
        department_name: this.getDepartmentName(formValue.department),
        settlement_name: this.getSettlementName(formValue.settlement),
        location_coordinates: formValue.location_coordinates?.lat && formValue.location_coordinates?.lng ? {
          lat: parseFloat(formValue.location_coordinates.lat),
          lng: parseFloat(formValue.location_coordinates.lng)
        } : undefined,
        pickup_available: formValue.pickup_available || false,
        delivery_available: formValue.delivery_available || false,
        tags: formValue.tags || [],
        is_active: formValue.is_active !== false
      };
      
      console.log('ProductData created:', productData);

      // Add category-specific details required by backend
      switch (mappedCategory) {
        case 'supplies':
          productData.supplies_details = {
            supply_type: 'seeds', // Basic supply type for seeds
            brand: null,
            model: null,
            active_ingredients: [],
            concentration: null,
            expiry_date: null,
            batch_number: null,
            registration_number: null,
            required_licenses: [],
            safety_data_sheet_url: null,
            storage_requirements: null,
            handling_instructions: null,
            disposal_instructions: null
          };
          break;
        case 'livestock':
          productData.livestock_details = {
            animal_type: 'cattle',
            breed: null,
            age_months: null,
            weight_kg: null,
            gender: null,
            health_certificates: [],
            vaccinations: null,
            last_veterinary_check: null,
            is_organic: false,
            is_pregnant: null,
            breeding_history: null,
            genetic_information: null
          };
          break;
        case 'transport':
          productData.transport_details = {
            vehicle_type: 'truck',
            capacity_tons: null,
            capacity_cubic_meters: null,
            price_per_km: null,
            has_refrigeration: false,
            has_livestock_equipment: false,
            service_provinces: [],
            min_distance_km: null,
            max_distance_km: null,
            license_plate: null,
            license_expiry: null,
            insurance_expiry: null,
            vehicle_year: null
          };
          break;
      }

      if (this.isEditMode && this.productId) {
        // For updates, send only basic product data that matches UpdateProductRequest
        const updateData: any = {
          title: productData.title,
          description: productData.description,
          price: productData.price,
          price_type: productData.price_type,
          province: productData.province,
          department: productData.department,
          settlement: productData.settlement,
          city: productData.city,
          // Add location names for display (now supported in UpdateProductRequest)
          province_name: this.getProvinceName(productData.province),
          department_name: this.getDepartmentName(productData.department),
          settlement_name: this.getSettlementName(productData.settlement),
          pickup_available: productData.pickup_available,
          delivery_available: productData.delivery_available,
          tags: productData.tags
          // Note: currency is still not in UpdateProductRequest
        };

        console.log('updateData before cleaning:', updateData);

        // Only add location_coordinates if valid
        if (productData.location_coordinates && 
            typeof productData.location_coordinates.lat === 'number' && 
            typeof productData.location_coordinates.lng === 'number' &&
            !isNaN(productData.location_coordinates.lat) && 
            !isNaN(productData.location_coordinates.lng)) {
          updateData.location_coordinates = productData.location_coordinates;
        }
        
        // Add category-specific details if they exist
        if (productData.livestock_details) {
          updateData.livestock_details = productData.livestock_details;
        }
        if (productData.transport_details) {
          updateData.transport_details = productData.transport_details;
        }
        
        // Check if we have new images to upload
        const newImageFiles: File[] = this.filePreviews
          .filter(preview => preview.type === 'image')
          .map(preview => preview.file);

        if (newImageFiles.length > 0) {
          // Use updateProductWithImages when we have new images
          this.productService.updateProductWithImages(this.productId, updateData, newImageFiles).subscribe({
            next: (response) => {
              console.log('Product updated with images successfully:', response);
              this.snackBar.open('Producto actualizado exitosamente con nuevas imágenes', 'Cerrar', { duration: 5000 });
              this.router.navigate(['/my-products']);
            },
            error: (error) => {
              console.error('Error updating product with images:', error);
              this.snackBar.open('Error al actualizar el producto con imágenes', 'Cerrar', { duration: 3000 });
            }
          });
        } else {
          // Use regular update method when no new images
          this.productService.updateProduct(this.productId, updateData).subscribe({
            next: (response) => {
              console.log('Product updated successfully:', response);
              this.snackBar.open('Producto actualizado exitosamente', 'Cerrar', { duration: 5000 });
              this.router.navigate(['/my-products']);
            },
            error: (error) => {
              console.error('Error updating product:', error);
              this.snackBar.open('Error al actualizar el producto', 'Cerrar', { duration: 3000 });
            }
          });
        }
      } else {
        // Extract image files from filePreviews
        const imageFiles: File[] = this.filePreviews
          .filter(preview => preview.type === 'image')
          .map(preview => preview.file);

        if (imageFiles.length > 0) {
          // Use FormData method for products with images
          this.productService.createProductWithImages(productData, imageFiles).subscribe({
            next: (response) => {
              this.snackBar.open('¡Producto creado exitosamente!', 'Cerrar', { duration: 3000 });
              this.router.navigate(['/my-products']);
            },
            error: (error) => {
              console.error('Error creating product:', error);
              this.snackBar.open('Error al crear el producto', 'Cerrar', { duration: 3000 });
            }
          });
        } else {
          // Use regular JSON method for products without images
          this.productService.createProduct(productData).subscribe({
            next: (response) => {
              this.snackBar.open('¡Producto creado exitosamente!', 'Cerrar', { duration: 3000 });
              this.router.navigate(['/my-products']);
            },
            error: (error) => {
              console.error('Error creating product:', error);
              this.snackBar.open('Error al crear el producto', 'Cerrar', { duration: 3000 });
            }
          });
        }
      }
    } else {
      this.snackBar.open('Por favor completa todos los campos requeridos', 'Cerrar', { duration: 3000 });
    }
  }

  removeImage(imageId: string): void {
    // For existing uploaded images, just remove from display - will be handled during update
    this.uploadedImages = this.uploadedImages.filter(img => img.id !== imageId);
    this.productForm.get('images')?.setValue(this.uploadedImages);
  }

  setPrimaryImage(imageId: string): void {
    // For existing uploaded images, just change locally - will be handled during update
    this.uploadedImages.forEach(img => {
      img.is_primary = img.id === imageId;
    });
  }

  updateImageOrder(imageId: string, newOrder: number): void {
    // For existing uploaded images, just change locally - will be handled during update
    const image = this.uploadedImages.find(img => img.id === imageId);
    if (image) {
      image.display_order = newOrder;
    }
    this.uploadedImages.sort((a, b) => a.display_order - b.display_order);
  }

  /**
   * Get the display URL for an existing uploaded image
   */
  getImageDisplayUrl(image: ProductImage): string {
    if (!image.image_url) {
      return '';
    }
    
    // The image_url from the API is already the correct URL, use it directly
    return image.image_url;
  }

  // Seller profile check methods
  private checkSellerProfileAndShowModal(): void {
    this.sellerService.checkSellerProfileComplete().subscribe({
      next: (response) => {
        if (!response.is_complete) {
          this.showSellerSetupModal();
        }
      },
      error: (error) => {
        console.error('Error checking seller profile:', error);
        // If there's an error checking, still show the modal to be safe
        this.showSellerSetupModal();
      }
    });
  }

  private showSellerSetupModal(): void {
    const dialogRef = this.dialog.open(SellerSetupModalComponent, {
      width: '90%',
      maxWidth: '700px',
      disableClose: true, // Prevent closing by clicking outside
      panelClass: 'seller-setup-dialog'
    });

    dialogRef.afterClosed().subscribe(result => {
      if (result) {
        // User completed the seller setup
        this.snackBar.open('¡Perfil configurado! Ahora puedes publicar productos', 'Cerrar', {
          duration: 3000,
          panelClass: ['success-snackbar']
        });
      } else {
        // User cancelled - redirect back to marketplace
        this.snackBar.open('Debes completar tu perfil de vendedor para publicar productos', 'Cerrar', {
          duration: 4000,
          panelClass: ['warning-snackbar']
        });
        this.router.navigate(['/marketplace']);
      }
    });
  }

  // Location methods
  loadLocationDataForEdit(provinceId: string, departmentId?: string, settlementId?: string): void {
    // Set the selected IDs
    this.selectedProvinceId = provinceId;
    this.selectedDepartmentId = departmentId || '';
    this.selectedSettlementId = settlementId || '';
    
    // Update form values
    this.productForm.patchValue({ 
      province: provinceId,
      department: departmentId || '',
      settlement: settlementId || ''
    });
    
    // Load departments for this province
    this.loadDepartments(provinceId);
    
    if (departmentId) {
      // Load settlements for this department
      this.loadSettlementsByDepartment(provinceId, departmentId);
    }
  }

  loadProvinces(): void {
    this.locationService.getProvinces({ max: 50 }).subscribe({
      next: (response) => {
        this.provinces = response.provincias;
      },
      error: (error) => {
        console.error('Error al cargar provincias:', error);
      }
    });
  }

  onProvinceChange(event: any): void {
    const provinceId = event.value || event.target?.value;
    this.selectedProvinceId = provinceId;
    
    // Clear dependent selects
    this.departments = [];
    this.settlements = [];
    this.selectedDepartmentId = '';
    this.selectedSettlementId = '';
    
    this.productForm.patchValue({
      department: '',
      settlement: ''
    });

    if (provinceId) {
      this.loadDepartments(provinceId);
      this.loadSettlements(provinceId);
    }
  }

  onDepartmentChange(event: any): void {
    const departmentId = event.value || event.target?.value;
    this.selectedDepartmentId = departmentId;
    
    this.settlements = [];
    this.productForm.patchValue({ settlement: '' });

    if (departmentId && this.selectedProvinceId) {
      this.loadSettlementsByDepartment(this.selectedProvinceId, departmentId);
    }
  }

  onSettlementChange(event: any): void {
    this.selectedSettlementId = event.value || event.target?.value;
  }

  loadDepartments(provinceId: string): void {
    this.locationService.getDepartments({ provincia: provinceId, max: 50 }).subscribe({
      next: (response) => {
        this.departments = response.departamentos || [];
      },
      error: (error) => {
        console.error('Error al cargar departamentos:', error);
        this.departments = [];
      }
    });
  }

  loadSettlements(provinceId: string): void {
    this.locationService.getSettlements({ provincia: provinceId, max: 100 }).subscribe({
      next: (response) => {
        this.settlements = response.asentamientos || [];
      },
      error: (error) => {
        console.error('Error al cargar asentamientos:', error);
        this.settlements = [];
      }
    });
  }

  loadSettlementsByDepartment(provinceId: string, departmentId: string): void {
    this.locationService.getSettlements({ provincia: provinceId, departamento: departmentId, max: 100 }).subscribe({
      next: (response) => {
        this.settlements = response.asentamientos || [];
      },
      error: (error) => {
        console.error('Error al cargar asentamientos por departamento:', error);
        this.settlements = [];
      }
    });
  }

  // Helper methods to get location names
  getProvinceName(provinceId: string): string {
    const province = this.provinces.find(p => p.id === provinceId);
    return province?.nombre || '';
  }

  getDepartmentName(departmentId: string): string {
    const department = this.departments.find(d => d.id === departmentId);
    return department?.nombre || '';
  }

  getSettlementName(settlementId: string): string {
    const settlement = this.settlements.find(s => s.id === settlementId);
    return settlement?.nombre || '';
  }
}