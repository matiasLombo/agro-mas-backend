import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { CommonModule } from '@angular/common';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { HttpClientModule, HTTP_INTERCEPTORS } from '@angular/common/http';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

// Angular Material
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSidenavModule } from '@angular/material/sidenav';
import { MatDividerModule } from '@angular/material/divider';
import { MatOptionModule } from '@angular/material/core';
import { MatSelectModule } from '@angular/material/select';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatDialogModule } from '@angular/material/dialog';
import { MatSnackBarModule } from '@angular/material/snack-bar';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatMenuModule } from '@angular/material/menu';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatChipsModule } from '@angular/material/chips';

// import { AngularFireModule } from '@angular/fire/compat';
// import { AngularFireStorageModule } from '@angular/fire/compat/storage';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';

// Auth Components
import { LoginComponent } from './auth/login/login.component';
import { RegisterComponent } from './auth/register/register.component';

// Marketplace Components
import { MarketplaceComponent } from './marketplace/marketplace.component';
// Product Components
import { ProductFormComponent } from './product-form/product-form.component';
// Seller Setup Modal
import { SellerSetupModalComponent } from './components/seller-setup-modal/seller-setup-modal.component';
import { ProductDetailComponent } from './product-detail/product-detail.component';
import { MyProductsComponent } from './my-products/my-products.component';
// Dialog Components
import { DeleteProductDialogComponent } from './components/delete-product-dialog/delete-product-dialog.component';

// Services
import { AuthService } from './core/services/auth.service';
import { HttpService } from './core/services/http.service';
import { ProductService } from './core/services/product.service';
import { SellerService } from './services/seller.service';

// Guards
import { AuthGuard } from './core/guards/auth.guard';
import { RoleGuard, SellerGuard } from './core/guards/role.guard';

// Interceptors
import { AuthInterceptor } from './core/interceptors/auth.interceptor';
import { ErrorInterceptor } from './core/interceptors/error.interceptor';

import { environment } from '@environments/environment';

@NgModule({
  declarations: [
    AppComponent,
    LoginComponent,
    RegisterComponent,
    MarketplaceComponent,
    ProductFormComponent,
    SellerSetupModalComponent,
    ProductDetailComponent,
    MyProductsComponent,
    DeleteProductDialogComponent
  ],
  imports: [
    BrowserModule,
    CommonModule,
    BrowserAnimationsModule,
    HttpClientModule,
    FormsModule,
    ReactiveFormsModule,
    AppRoutingModule,
    MatIconModule,
    MatButtonModule,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatSidenavModule,
    MatDividerModule,
    MatSelectModule,
    MatOptionModule,
    MatProgressBarModule,
    MatDialogModule,
    MatSnackBarModule,
    MatToolbarModule,
    MatMenuModule,
    MatTooltipModule,
    MatProgressSpinnerModule,
    MatChipsModule,

    // Configuración de Firebase (temporalmente deshabilitado)
    // AngularFireModule.initializeApp(environment.firebaseConfig),
    // AngularFireStorageModule

  ],
  providers: [
    AuthService,
    HttpService,
    ProductService,
    SellerService,
    AuthGuard,
    RoleGuard,
    SellerGuard,
    {
      provide: HTTP_INTERCEPTORS,
      useClass: AuthInterceptor,
      multi: true
    },
    {
      provide: HTTP_INTERCEPTORS,
      useClass: ErrorInterceptor,
      multi: true
    }
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
