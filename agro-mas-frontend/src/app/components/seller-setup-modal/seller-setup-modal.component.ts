import { Component, OnInit, Inject } from '@angular/core';
import { FormBuilder, FormGroup, Validators, AbstractControl, ValidationErrors } from '@angular/forms';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { SellerService, SellerProfileRequest } from '../../services/seller.service';
import { AuthService } from '../../core/services/auth.service';
import { provinces } from '../../core/data/argentina.data';

@Component({
  selector: 'app-seller-setup-modal',
  templateUrl: './seller-setup-modal.component.html',
  styleUrls: ['./seller-setup-modal.component.scss']
})
export class SellerSetupModalComponent implements OnInit {
  sellerForm: FormGroup;
  provinces = provinces;
  isSubmitting = false;
  isLoading = true;
  isFirstTimeSetup = true;

  constructor(
    private fb: FormBuilder,
    private sellerService: SellerService,
    private authService: AuthService,
    private snackBar: MatSnackBar,
    public dialogRef: MatDialogRef<SellerSetupModalComponent>,
    @Inject(MAT_DIALOG_DATA) public data: any
  ) {
    this.sellerForm = this.createForm();
  }

  ngOnInit(): void {
    // Load existing profile data if available
    this.loadExistingProfile();
  }

  private createForm(): FormGroup {
    return this.fb.group({
      business_name: ['', [Validators.required, Validators.minLength(2)]],
      cuit: ['', [Validators.required]],
      cbu: ['', [Validators.required]],
      cbu_alias: [''],
      bank_name: [''],
      renspa: ['', [Validators.required]],
      establishment_name: [''],
      establishment_location: [''],
      business_type: ['']
    });
  }

  private loadExistingProfile(): void {
    this.sellerService.getSellerProfile().subscribe({
      next: (response) => {
        if (response.profile) {
          // Si hay datos existentes, no es la primera configuración
          this.isFirstTimeSetup = false;
          
          this.sellerForm.patchValue({
            business_name: response.profile.business_name || '',
            cuit: response.profile.cuit || '',
            cbu: response.profile.cbu || '',
            cbu_alias: response.profile.cbu_alias || '',
            bank_name: response.profile.bank_name || '',
            renspa: response.profile.renspa || '',
            establishment_name: response.profile.establishment_name || '',
            establishment_location: response.profile.establishment_location || '',
            business_type: response.profile.business_type || ''
          });
        }
        this.isLoading = false;
      },
      error: (error) => {
        console.log('No existing profile found, starting fresh');
        // Si hay error, mantenemos isFirstTimeSetup = true
        this.isLoading = false;
      }
    });
  }

  onSubmit(): void {
    if (this.sellerForm.valid && !this.isSubmitting) {
      this.isSubmitting = true;
      
      const currentUser = this.authService.currentUser;
      if (!currentUser) {
        this.snackBar.open('Error: No se pudo obtener información del usuario', 'Cerrar', {
          duration: 5000,
          panelClass: ['error-snackbar']
        });
        this.isSubmitting = false;
        return;
      }

      const profileData: SellerProfileRequest = {
        ...this.sellerForm.value,
        // Add required contact information from current user
        phone: currentUser.phone || '',
        address: currentUser.address || '',
        city: currentUser.city || '',
        province: currentUser.province || ''
      };
      
      this.sellerService.upgradeToSeller(profileData).subscribe({
        next: (response) => {
          const message = this.isFirstTimeSetup 
            ? '🎉 ¡Perfil de vendedor configurado correctamente!'
            : 'Información actualizada correctamente';
            
          this.snackBar.open(message, 'Cerrar', {
            duration: 5000,
            panelClass: ['success-snackbar']
          });
          this.dialogRef.close(true);
        },
        error: (error) => {
          this.isSubmitting = false;
          let errorMessage = 'Error al configurar el perfil de vendedor';
          
          if (error.error?.code === 'CUIT_EXISTS') {
            errorMessage = 'Este CUIT ya está registrado por otro usuario';
          } else if (error.error?.message) {
            errorMessage = error.error.message;
          }
          
          this.snackBar.open(errorMessage, 'Cerrar', {
            duration: 5000,
            panelClass: ['error-snackbar']
          });
        }
      });
    }
  }

  onCancel(): void {
    this.dialogRef.close(false);
  }

  // Helper methods for form validation
  getFieldError(fieldName: string): string {
    const field = this.sellerForm.get(fieldName);
    if (field?.errors && field.touched) {
      if (field.errors['required']) return `${this.getFieldDisplayName(fieldName)} es requerido`;
      if (field.errors['minLength']) return `${this.getFieldDisplayName(fieldName)} es muy corto`;
    }
    return '';
  }

  private getFieldDisplayName(fieldName: string): string {
    const displayNames: {[key: string]: string} = {
      'business_name': 'Nombre comercial',
      'cuit': 'CUIT',
      'cbu': 'CBU/CVU',
      'renspa': 'Código RENSPA',
    };
    return displayNames[fieldName] || fieldName;
  }


}