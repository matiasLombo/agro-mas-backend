import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { HttpService } from '../core/services/http.service';

export interface SellerProfile {
  cuit?: string;
  cbu?: string;
  cbu_alias?: string;
  bank_name?: string;
  renspa?: string;
  establishment_name?: string;
  establishment_location?: string;
  business_name?: string;
  business_type?: string;
  phone?: string;
  address?: string;
  city?: string;
  province?: string;
  is_complete: boolean;
}

export interface SellerProfileRequest {
  cuit: string;
  cbu: string;
  cbu_alias?: string;
  bank_name?: string;
  renspa?: string;
  establishment_name?: string;
  establishment_location?: string;
  business_name: string;
  business_type?: string;
  phone: string;
  address: string;
  city: string;
  province: string;
}

@Injectable({
  providedIn: 'root'
})
export class SellerService {

  constructor(private http: HttpService) {}

  /**
   * Get current seller profile status
   */
  getSellerProfile(): Observable<{profile: SellerProfile}> {
    return this.http.get<{profile: SellerProfile}>('/auth/seller-profile');
  }

  /**
   * Check if seller profile is complete
   */
  checkSellerProfileComplete(): Observable<{is_complete: boolean}> {
    return this.http.get<{is_complete: boolean}>('/auth/seller-profile/check');
  }

  /**
   * Upgrade user to seller with complete profile
   */
  upgradeToSeller(profile: SellerProfileRequest): Observable<{message: string}> {
    return this.http.post<{message: string}>('/auth/upgrade-to-seller', profile);
  }
}