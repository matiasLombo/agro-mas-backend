export interface Province {
    id: string;
    nombre: string;
    centroide: {
        lat: number;
        lon: number;
    };
}

export interface Department {
    id: string;
    nombre: string;
    centroide: {
        lat: number;
        lon: number;
    };
    provincia: {
        id: string;
        nombre: string;
    };
}

export interface Municipality {
    id: string;
    nombre: string;
    centroide: {
        lat: number;
        lon: number;
    };
    provincia: {
        id: string;
        nombre: string;
    };
}

export interface Locality {
    id: string;
    nombre: string;
    centroide: {
        lat: number;
        lon: number;
    };
    provincia: {
        id: string;
        nombre: string;
    };
    municipio?: {
        id: string;
        nombre: string;
    };
}

export interface Settlement {
    id: string;
    nombre: string;
    centroide: {
        lat: number;
        lon: number;
    };
    provincia: {
        id: string;
        nombre: string;
    };
    departamento?: {
        id: string;
        nombre: string;
    };
}

export interface LocationResponse {
    lat: number;
    lon: number;
    provincia: Province;
    municipio?: Municipality;
    localidad?: Locality;
}