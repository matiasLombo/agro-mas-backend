import { Injectable } from '@angular/core';
import { HttpErrorResponse, HttpInterceptor, HttpRequest, HttpHandler } from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError } from 'rxjs/operators';

@Injectable()
export class ErrorInterceptor implements HttpInterceptor {
  
  intercept(request: HttpRequest<any>, next: HttpHandler): Observable<any> {
    return next.handle(request).pipe(
      catchError((error: HttpErrorResponse) => {
        let errorMessage = 'Ha ocurrido un error inesperado';
        
        if (error.error?.code) {
          // Handle specific error codes from backend
          switch (error.error.code) {
            case 'USER_EXISTS':
              errorMessage = 'Ya existe una cuenta con este correo electrónico. ¿Quieres iniciar sesión en su lugar?';
              break;
            case 'CUIT_EXISTS':
              errorMessage = 'Este CUIT ya está registrado por otro usuario. Por favor, verifica los datos ingresados.';
              break;
            case 'INVALID_CREDENTIALS':
            case 'USER_NOT_FOUND':
              errorMessage = 'Correo electrónico o contraseña incorrectos';
              break;
            case 'ACCOUNT_INACTIVE':
              errorMessage = 'Tu cuenta ha sido desactivada. Contacta al soporte para más información.';
              break;
            case 'REGISTRATION_FAILED':
              errorMessage = 'No se pudo crear la cuenta. Verifica que todos los datos sean correctos.';
              break;
            case 'INVALID_REQUEST':
              errorMessage = 'Los datos proporcionados no son válidos. Revisa la información ingresada.';
              break;
            case 'SELLER_UPGRADE_FAILED':
              errorMessage = 'No se pudo configurar tu perfil de vendedor. Verifica que todos los campos estén completos.';
              break;
            case 'AUTH_REQUIRED':
              errorMessage = 'Necesitas iniciar sesión para realizar esta acción';
              break;
            default:
              if (error.error.message) {
                errorMessage = error.error.message;
              }
              break;
          }
        } else if (error.status === 0) {
          errorMessage = 'No se pudo conectar con el servidor. Verifica tu conexión a internet.';
        } else if (error.status >= 500) {
          errorMessage = 'Error interno del servidor. Intenta nuevamente en unos momentos.';
        } else if (error.status === 404) {
          errorMessage = 'El recurso solicitado no fue encontrado.';
        } else if (error.status === 401) {
          errorMessage = 'No tienes autorización para realizar esta acción.';
        } else if (error.status === 403) {
          errorMessage = 'No tienes permisos para acceder a este recurso.';
        }

        // Return the error with the friendly message
        return throwError(() => ({
          ...error,
          message: errorMessage,
          originalError: error
        }));
      })
    );
  }
}