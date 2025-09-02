import { Injectable } from '@angular/core';
import { CanActivate, ActivatedRouteSnapshot, RouterStateSnapshot, Router } from '@angular/router';
import { Observable, map, take } from 'rxjs';
import { AuthService } from '../services/auth.service';

@Injectable({
  providedIn: 'root'
})
export class RoleGuard implements CanActivate {
  constructor(
    private authService: AuthService,
    private router: Router
  ) {}

  canActivate(
    route: ActivatedRouteSnapshot,
    state: RouterStateSnapshot
  ): Observable<boolean> {
    return this.authService.isAuthenticated$.pipe(
      take(1),
      map(isAuthenticated => {
        if (!isAuthenticated) {
          this.router.navigate(['/login']);
          return false;
        }

        const allowedRoles = route.data?.['allowedRoles'] as string[];
        const userRole = this.authService.currentUser?.role;

        if (allowedRoles && userRole && allowedRoles.includes(userRole)) {
          return true;
        } else {
          this.router.navigate(['/marketplace'], {
            queryParams: { message: 'No tienes permisos para acceder a esta sección' }
          });
          return false;
        }
      })
    );
  }
}

@Injectable({
  providedIn: 'root'
})
export class SellerGuard implements CanActivate {
  constructor(
    private authService: AuthService,
    private router: Router
  ) {}

  canActivate(): Observable<boolean> {
    return this.authService.isAuthenticated$.pipe(
      take(1),
      map(isAuthenticated => {
        if (!isAuthenticated) {
          this.router.navigate(['/login']);
          return false;
        }

        if (this.authService.isSeller() || this.authService.isAdmin()) {
          return true;
        } else {
          this.router.navigate(['/marketplace'], {
            queryParams: { message: 'Solo vendedores pueden acceder a esta sección' }
          });
          return false;
        }
      })
    );
  }
}