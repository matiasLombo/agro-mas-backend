import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, BehaviorSubject, tap, map } from 'rxjs';
import { User, LoginRequest, RegisterRequest, AuthResponse } from '../models/user.model';
import { environment } from '../../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private currentUserSubject = new BehaviorSubject<User | null>(null);
  public currentUser$ = this.currentUserSubject.asObservable();

  private isAuthenticatedSubject = new BehaviorSubject<boolean>(false);
  public isAuthenticated$ = this.isAuthenticatedSubject.asObservable();

  constructor(private http: HttpClient) {
    this.loadStoredAuth();
  }

  get currentUser(): User | null {
    return this.currentUserSubject.value;
  }

  get isAuthenticated(): boolean {
    return this.isAuthenticatedSubject.value;
  }

  login(credentials: LoginRequest): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${environment.apiUrl}/auth/login`, credentials)
      .pipe(
        tap(response => {
          this.saveAuthData(response);
        })
      );
  }

  register(userData: RegisterRequest): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${environment.apiUrl}/auth/register`, userData)
      .pipe(
        tap(response => {
          this.saveAuthData(response);
        })
      );
  }

  logout(): void {
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
    localStorage.removeItem('currentUser');
    this.currentUserSubject.next(null);
    this.isAuthenticatedSubject.next(false);
  }

  refreshToken(): Observable<AuthResponse> {
    const refreshToken = localStorage.getItem('refreshToken');
    return this.http.post<AuthResponse>(`${environment.apiUrl}/auth/refresh`, { refreshToken })
      .pipe(
        tap(response => {
          this.saveAuthData(response);
        })
      );
  }

  private saveAuthData(authResponse: AuthResponse): void {
    if (authResponse.token) {
      // The backend returns TokenResponse directly as 'token'
      localStorage.setItem('accessToken', authResponse.token.access_token);
      localStorage.setItem('refreshToken', authResponse.token.refresh_token);
    }
    if (authResponse.user) {
      localStorage.setItem('currentUser', JSON.stringify(authResponse.user));
      this.currentUserSubject.next(authResponse.user);
      this.isAuthenticatedSubject.next(true);
    }
  }

  private loadStoredAuth(): void {
    const token = localStorage.getItem('accessToken');
    const userStr = localStorage.getItem('currentUser');
    
    if (token && userStr) {
      try {
        const user = JSON.parse(userStr);
        this.currentUserSubject.next(user);
        this.isAuthenticatedSubject.next(true);
      } catch (error) {
        this.logout();
      }
    }
  }

  // Role checking methods
  isSeller(): boolean {
    return this.currentUser?.role === 'seller';
  }

  isBuyer(): boolean {
    return this.currentUser?.role === 'buyer';
  }

  isAdmin(): boolean {
    return this.currentUser?.role === 'admin';
  }

  hasRole(role: string): boolean {
    return this.currentUser?.role === role;
  }

  hasMinVerificationLevel(level: number): boolean {
    return (this.currentUser?.verification_level || 0) >= level;
  }

  updateProfile(profileData: Partial<User>): Observable<User> {
    return this.http.put<any>(`${environment.apiUrl}/auth/profile`, profileData)
      .pipe(
        tap(response => {
          // Backend returns {message: string, user: User}
          const updatedUser = response.user;
          // Update the current user in memory and localStorage
          localStorage.setItem('currentUser', JSON.stringify(updatedUser));
          this.currentUserSubject.next(updatedUser);
        }),
        // Return just the user object for consistency
        map(response => response.user)
      );
  }
}