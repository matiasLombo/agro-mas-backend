import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders, HttpErrorResponse } from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError, retry } from 'rxjs/operators';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class HttpService {
  private readonly baseUrl = environment.apiUrl;

  constructor(private http: HttpClient) {}

  // GET request
  get<T>(endpoint: string): Observable<T> {
    return this.http.get<T>(`${this.baseUrl}${endpoint}`, {
      headers: this.getHeaders()
    }).pipe(
      retry(1),
      catchError(this.handleError)
    );
  }

  // POST request
  post<T>(endpoint: string, data: any): Observable<T> {
    return this.http.post<T>(`${this.baseUrl}${endpoint}`, data, {
      headers: this.getHeaders()
    }).pipe(
      catchError(this.handleError)
    );
  }

  // PUT request
  put<T>(endpoint: string, data: any): Observable<T> {
    return this.http.put<T>(`${this.baseUrl}${endpoint}`, data, {
      headers: this.getHeaders()
    }).pipe(
      catchError(this.handleError)
    );
  }

  // DELETE request
  delete<T>(endpoint: string): Observable<T> {
    return this.http.delete<T>(`${this.baseUrl}${endpoint}`, {
      headers: this.getHeaders()
    }).pipe(
      catchError(this.handleError)
    );
  }

  private getHeaders(): HttpHeaders {
    const headers: any = {
      'Content-Type': 'application/json',
      'Accept': 'application/json'
    };

    // Add auth token if available
    const token = localStorage.getItem('authToken');
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    return new HttpHeaders(headers);
  }

  private handleError = (error: HttpErrorResponse): Observable<never> => {
    let errorMessage = 'Ha ocurrido un error inesperado';

    if (error.error instanceof ErrorEvent) {
      // Client-side error
      errorMessage = `Error: ${error.error.message}`;
    } else {
      // Server-side error
      switch (error.status) {
        case 400:
          errorMessage = error.error?.message || 'Solicitud inválida';
          break;
        case 401:
          errorMessage = 'No autorizado. Por favor, inicia sesión nuevamente';
          // Clear token and redirect to login
          localStorage.removeItem('authToken');
          break;
        case 403:
          errorMessage = 'No tienes permisos para realizar esta acción';
          break;
        case 404:
          errorMessage = 'Recurso no encontrado';
          break;
        case 500:
          errorMessage = 'Error interno del servidor. Intenta nuevamente más tarde';
          break;
        default:
          errorMessage = `Error ${error.status}: ${error.message}`;
      }
    }

    console.error('HTTP Error:', error);
    return throwError(() => new Error(errorMessage));
  }

  // Health check method
  healthCheck(): Observable<any> {
    const healthUrl = `${this.baseUrl.replace('/api/v1', '')}/health`;
    return this.http.get(healthUrl).pipe(
      catchError(this.handleError)
    );
  }

  // Environment check method
  environmentCheck(): Observable<any> {
    const envUrl = `${this.baseUrl.replace('/api/v1', '')}/env`;
    return this.http.get(envUrl).pipe(
      catchError(this.handleError)
    );
  }
}
