import { AbstractControl, ValidatorFn, ValidationErrors } from '@angular/forms';

export function cuitValidator(): ValidatorFn {
  return (control: AbstractControl): ValidationErrors | null => {
    const value = control.value;
    
    if (!value) {
      return null; // No validar si está vacío
    }

    // Limpiar formato (remover guiones y espacios)
    const cuit = value.replace(/[-\s]/g, '');
    
    // Verificar longitud
    if (cuit.length !== 11) {
      return { cuit: { message: 'CUIT debe tener 11 dígitos' } };
    }

    // Verificar que sean solo números
    if (!/^\d{11}$/.test(cuit)) {
      return { cuit: { message: 'CUIT debe contener solo números' } };
    }

    // Algoritmo de verificación CUIT
    const coeficientes = [5, 4, 3, 2, 7, 6, 5, 4, 3, 2];
    const digitos = cuit.split('').map((d: string) => parseInt(d));
    
    let suma = 0;
    for (let i = 0; i < 10; i++) {
      suma += digitos[i] * coeficientes[i];
    }
    
    let digitoVerificador = 11 - (suma % 11);
    if (digitoVerificador === 11) {
      digitoVerificador = 0;
    } else if (digitoVerificador === 10) {
      return { cuit: { message: 'CUIT inválido' } };
    }
    
    if (digitoVerificador !== digitos[10]) {
      return { cuit: { message: 'CUIT inválido - dígito verificador incorrecto' } };
    }

    // Verificar tipo de persona según primeros dígitos
    const tipoPersona = cuit.substring(0, 2);
    const tiposValidos = ['20', '23', '24', '27', '30', '33', '34'];
    
    if (!tiposValidos.includes(tipoPersona)) {
      return { cuit: { message: 'Tipo de CUIT no válido' } };
    }

    return null;
  };
}

export function formatCuit(cuit: string): string {
  if (!cuit) return '';
  
  const clean = cuit.replace(/[-\s]/g, '');
  if (clean.length !== 11) return cuit;
  
  return `${clean.substring(0, 2)}-${clean.substring(2, 10)}-${clean.substring(10)}`;
}

export function getCuitEntityType(cuit: string): string {
  if (!cuit) return '';
  
  const clean = cuit.replace(/[-\s]/g, '');
  if (clean.length !== 11) return '';
  
  const tipo = clean.substring(0, 2);
  
  const tipos: { [key: string]: string } = {
    '20': 'Persona Física Masculino',
    '23': 'Persona Física Masculino',
    '24': 'Persona Física Masculino',
    '27': 'Persona Física Femenino',
    '30': 'Persona Jurídica',
    '33': 'Persona Jurídica',
    '34': 'Persona Jurídica'
  };
  
  return tipos[tipo] || 'Desconocido';
}