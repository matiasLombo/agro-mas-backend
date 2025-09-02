import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Province, Department, Municipality, Locality, Settlement, LocationResponse } from '@core/models/location.model';

@Injectable({
  providedIn: 'root'
})
export class LocationService {
  private apiUrl = 'https://apis.datos.gob.ar/georef/api';

  constructor(private http: HttpClient) { }

  /**
   * Get all provinces
   */
  getProvinces(params?: {
    nombre?: string;
    campos?: string[];
    max?: number;
  }): Observable<{ provincias: Province[] }> {
    let httpParams = new HttpParams();

    if (params?.nombre) httpParams = httpParams.set('nombre', params.nombre);
    if (params?.campos && params.campos.length > 0) httpParams = httpParams.set('campos', params.campos.join(','));
    if (params?.max) httpParams = httpParams.set('max', params.max.toString());

    return this.http.get<{ provincias: Province[] }>(`${this.apiUrl}/provincias`, { params: httpParams });
  }

  /**
   * Get departments by province
   */
  getDepartments(params?: {
    provincia?: string;
    nombre?: string;
    campos?: string[];
    max?: number;
  }): Observable<{ departamentos: Department[] }> {
    let httpParams = new HttpParams();

    if (params?.provincia) httpParams = httpParams.set('provincia', params.provincia);
    if (params?.nombre) httpParams = httpParams.set('nombre', params.nombre);
    if (params?.campos && params.campos.length > 0) httpParams = httpParams.set('campos', params.campos.join(','));
    if (params?.max) httpParams = httpParams.set('max', params.max.toString());

    return this.http.get<{ departamentos: Department[] }>(`${this.apiUrl}/departamentos`, { params: httpParams });
  }

  /**
   * Get municipalities by province
   */
  getMunicipalities(params?: {
    provincia?: string;
    nombre?: string;
    campos?: string[];
    max?: number;
  }): Observable<{ municipios: Municipality[] }> {
    let httpParams = new HttpParams();

    if (params?.provincia) httpParams = httpParams.set('provincia', params.provincia);
    if (params?.nombre) httpParams = httpParams.set('nombre', params.nombre);
    if (params?.campos && params.campos.length > 0) httpParams = httpParams.set('campos', params.campos.join(','));
    if (params?.max) httpParams = httpParams.set('max', params.max.toString());

    return this.http.get<{ municipios: Municipality[] }>(`${this.apiUrl}/municipios`, { params: httpParams });
  }

  /**
   * Get localities by province and/or municipality
   */
  getLocalities(params?: {
    provincia?: string;
    municipio?: string;
    nombre?: string;
    campos?: string[];
    max?: number;
  }): Observable<{ localidades: Locality[] }> {
    let httpParams = new HttpParams();

    if (params?.provincia) httpParams = httpParams.set('provincia', params.provincia);
    if (params?.municipio) httpParams = httpParams.set('municipio', params.municipio);
    if (params?.nombre) httpParams = httpParams.set('nombre', params.nombre);
    if (params?.campos && params.campos.length > 0) httpParams = httpParams.set('campos', params.campos.join(','));
    if (params?.max) httpParams = httpParams.set('max', params.max.toString());

    return this.http.get<{ localidades: Locality[] }>(`${this.apiUrl}/localidades`, { params: httpParams });
  }

  /**
   * Get location information from coordinates
   */
  getLocationByCoordinates(lat: number, lon: number): Observable<LocationResponse> {
    const params = new HttpParams()
      .set('lat', lat.toString())
      .set('lon', lon.toString());

    return this.http.get<LocationResponse>(`${this.apiUrl}/ubicacion`, { params });
  }

  /**
   * Get settlements by province and/or department
   */
  getSettlements(params?: {
    provincia?: string;
    departamento?: string;
    nombre?: string;
    campos?: string[];
    max?: number;
  }): Observable<{ asentamientos: Settlement[] }> {
    let httpParams = new HttpParams();
    
    if (params?.provincia) httpParams = httpParams.set('provincia', params.provincia);
    if (params?.departamento) httpParams = httpParams.set('departamento', params.departamento);
    if (params?.nombre) httpParams = httpParams.set('nombre', params.nombre);
    if (params?.campos && params.campos.length > 0) httpParams = httpParams.set('campos', params.campos.join(','));
    if (params?.max) httpParams = httpParams.set('max', params.max.toString());

    return this.http.get<{ asentamientos: Settlement[] }>(`${this.apiUrl}/asentamientos`, { params: httpParams });
  }

  /**
   * Normalize address or location name
   */
  normalizeAddress(direccion: string): Observable<any> {
    const params = new HttpParams().set('direccion', direccion);

    return this.http.get<any>(`${this.apiUrl}/direcciones`, { params });
  }
}