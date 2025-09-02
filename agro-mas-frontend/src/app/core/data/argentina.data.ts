export interface Region {
  id: string;
  name: string;
  cities: City[];
}

export interface City {
  id: string;
  name: string;
}

export const REGIONS: Region[] = [
  {
    id: 'norte',
    name: 'Región Norte',
    cities: [
      { id: 'ciudad-norte-1', name: 'Ciudad Norte Principal' },
      { id: 'ciudad-norte-2', name: 'Ciudad Norte Industrial' },
      { id: 'ciudad-norte-3', name: 'Ciudad Norte Agrícola' }
    ]
  },
  {
    id: 'centro',
    name: 'Región Centro',
    cities: [
      { id: 'ciudad-centro-1', name: 'Ciudad Centro Principal' },
      { id: 'ciudad-centro-2', name: 'Ciudad Centro Comercial' },
      { id: 'ciudad-centro-3', name: 'Ciudad Centro Rural' }
    ]
  },
  {
    id: 'sur',
    name: 'Región Sur',
    cities: [
      { id: 'ciudad-sur-1', name: 'Ciudad Sur Principal' },
      { id: 'ciudad-sur-2', name: 'Ciudad Sur Ganadera' },
      { id: 'ciudad-sur-3', name: 'Ciudad Sur Costera' }
    ]
  },
  {
    id: 'este',
    name: 'Región Este',
    cities: [
      { id: 'ciudad-este-1', name: 'Ciudad Este Principal' },
      { id: 'ciudad-este-2', name: 'Ciudad Este Fronteriza' },
      { id: 'ciudad-este-3', name: 'Ciudad Este Rural' }
    ]
  },
  {
    id: 'oeste',
    name: 'Región Oeste',
    cities: [
      { id: 'ciudad-oeste-1', name: 'Ciudad Oeste Principal' },
      { id: 'ciudad-oeste-2', name: 'Ciudad Oeste Montañosa' },
      { id: 'ciudad-oeste-3', name: 'Ciudad Oeste Minera' }
    ]
  },
  {
    id: 'andina',
    name: 'Región Andina',
    cities: [
      { id: 'ciudad-andina-1', name: 'Ciudad Andina Principal' },
      { id: 'ciudad-andina-2', name: 'Ciudad Andina Valle' },
      { id: 'ciudad-andina-3', name: 'Ciudad Andina Alta' }
    ]
  }
];

// Provincias de Argentina
export const provinces = [
  'Buenos Aires',
  'Catamarca',
  'Chaco',
  'Chubut',
  'Córdoba',
  'Corrientes',
  'Entre Ríos',
  'Formosa',
  'Jujuy',
  'La Pampa',
  'La Rioja',
  'Mendoza',
  'Misiones',
  'Neuquén',
  'Río Negro',
  'Salta',
  'San Juan',
  'San Luis',
  'Santa Cruz',
  'Santa Fe',
  'Santiago del Estero',
  'Tierra del Fuego',
  'Tucumán',
  'Ciudad Autónoma de Buenos Aires'
];

// Alias para compatibilidad con código existente
export const ARGENTINA_PROVINCES = REGIONS;
export type Province = Region;