import { Component, OnInit, ChangeDetectorRef } from '@angular/core';
import { Router } from '@angular/router';
import { MatDialog } from '@angular/material/dialog';
import { ApiTestService } from './services/api-test.service';
import { AuthService } from './core/services/auth.service';
import { User } from './core/models/user.model';
import { Observable } from 'rxjs';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent implements OnInit {
  title = 'Agro Mas';
  backendStatus = 'Verificando...';
  backendConnected = false;
  backendEnvironment = '';
  environmentInfo: any = null;

  // Auth properties
  isAuthenticated$: Observable<boolean>;
  currentUser$: Observable<User | null>;
  
  // User menu state
  showUserMenu = false;

  constructor(
    private apiTestService: ApiTestService,
    private authService: AuthService,
    private router: Router,
    private dialog: MatDialog,
    private cdr: ChangeDetectorRef
  ) {
    this.isAuthenticated$ = this.authService.isAuthenticated$;
    this.currentUser$ = this.authService.currentUser$;
  }

  ngOnInit() {
    this.testBackendConnection();
    this.testBackendEnvironment();
    
    // Close user menu when clicking outside
    document.addEventListener('click', (event) => {
      const target = event.target as HTMLElement;
      const userMenuContainer = document.querySelector('.user-menu-container');
      
      if (userMenuContainer && !userMenuContainer.contains(target)) {
        this.showUserMenu = false;
        this.cdr.detectChanges();
      }
    });
  }

  get isAuthenticated(): boolean {
    return this.authService.isAuthenticated;
  }

  get currentUser(): User | null {
    return this.authService.currentUser;
  }

  logout(): void {
    this.authService.logout();
  }

  // User menu methods
  openEditProfile(): void {
    console.log('Opening edit profile...');
    import('./components/edit-profile-modal/edit-profile-modal.component').then(m => {
      const dialogRef = this.dialog.open(m.EditProfileModalComponent, {
        width: '700px',
        maxWidth: '95vw',
        maxHeight: '90vh',
        disableClose: false,
        hasBackdrop: true,
        backdropClass: 'custom-backdrop',
        panelClass: 'custom-dialog-container'
      });

      dialogRef.afterClosed().subscribe(result => {
        if (result && result.updated) {
          console.log('Profile updated successfully:', result.data);
          // Here you could update the local user data or refresh from the server
        }
      });
    });
  }

  openAccountSettings(): void {
    console.log('Opening account settings...');
    import('./components/seller-setup-modal/seller-setup-modal.component').then(m => {
      const dialogRef = this.dialog.open(m.SellerSetupModalComponent, {
        width: '800px',
        maxWidth: '95vw',
        maxHeight: '90vh',
        data: { isEdit: true },
        hasBackdrop: true,
        backdropClass: 'custom-backdrop',
        panelClass: 'custom-dialog-container'
      });

      dialogRef.afterClosed().subscribe(result => {
        if (result) {
          console.log('Account configuration updated successfully');
          // Here you could update the local user settings or refresh from the server
        }
      });
    });
  }

  openSellerSettings(): void {
    console.log('Opening seller settings...');
    // Open the seller setup modal for editing
    import('./components/seller-setup-modal/seller-setup-modal.component').then(m => {
      const dialogRef = this.dialog.open(m.SellerSetupModalComponent, {
        width: '800px',
        maxWidth: '95vw',
        data: { isEdit: true }
      });

      dialogRef.afterClosed().subscribe(result => {
        if (result) {
          console.log('Seller profile updated successfully');
        }
      });
    });
  }

  viewMyActivity(): void {
    console.log('Viewing activity...');
    this.router.navigate(['/my-purchases']); // Navigate to activity page
  }

  // User menu methods
  toggleUserMenu(): void {
    this.showUserMenu = !this.showUserMenu;
    this.cdr.detectChanges();
  }

  closeUserMenu(): void {
    this.showUserMenu = false;
    this.cdr.detectChanges();
  }

  testBackendConnection() {
    this.apiTestService.testBackendConnection().subscribe({
      next: (response) => {
        this.backendStatus = '✅ Backend conectado correctamente';
        this.backendConnected = true;
        console.log('Backend health:', response);
      },
      error: (error) => {
        this.backendStatus = '❌ Error conectando con backend (CORS/Network)';
        this.backendConnected = false;
        console.error('Backend connection error:', error);
      }
    });
  }

  testBackendEnvironment() {
    this.apiTestService.testBackendEnvironment().subscribe({
      next: (response) => {
        this.backendEnvironment = response.environment || 'Unknown';
        this.environmentInfo = response;
        console.log('Backend environment:', response);
      },
      error: (error) => {
        this.backendEnvironment = 'Error';
        console.error('Backend environment error:', error);
      }
    });
  }
}
