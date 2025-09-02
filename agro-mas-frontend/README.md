# Agro Mas Frontend

Frontend Angular para el marketplace agropecuario Agro Mas, diseñado para conectar productores agropecuarios con compradores directos.

## 🌾 Características

- **Marketplace Agropecuario**: Plataforma para compra/venta de productos agrícolas
- **Autenticación JWT**: Sistema de login seguro con roles (comprador/vendedor/admin)  
- **Gestión de Productos**: CRUD completo de productos con imágenes y geolocalización
- **Transacciones**: Sistema de pedidos, inquiries y reviews
- **Responsive Design**: Diseño adaptativo con Angular Material
- **PWA Ready**: Service Worker configurado para funcionalidad offline

## 🚀 Tecnologías

- **Angular 17+** con Standalone Components
- **Angular Material** para UI/UX
- **RxJS** para programación reactiva
- **Leaflet** para mapas interactivos
- **Firebase Hosting** para deployment
- **TypeScript** con configuración estricta

## 📦 Instalación

```bash
# Clonar el repositorio
git clone https://github.com/tu-usuario/agro-mas-frontend.git
cd agro-mas-frontend

# Instalar dependencias
npm install

# Configurar variables de entorno
cp src/environments/environment.ts.example src/environments/environment.ts
cp src/environments/environment.prod.ts.example src/environments/environment.prod.ts

# Ejecutar en desarrollo
npm start
```

## ⚙️ Configuración

### Variables de Entorno

Edita `src/environments/environment.ts`:

```typescript
export const environment = {
  production: false,
  apiUrl: 'http://localhost:8080/api/v1', // URL del backend Go
  firebaseConfig: {
    apiKey: 'tu-firebase-api-key',
    authDomain: 'agro-mas-dev.firebaseapp.com',
    projectId: 'agro-mas-dev',
    // ... resto de configuración Firebase
  }
};
```

### Firebase Setup

```bash
# Instalar Firebase CLI
npm install -g firebase-tools

# Login a Firebase
firebase login

# Inicializar proyecto (ya configurado)
firebase init hosting

# Deploy
npm run deploy
```

## 🏗️ Arquitectura

```
src/
├── app/
│   ├── core/                 # Servicios core, guards, interceptors
│   │   ├── models/          # Interfaces TypeScript
│   │   ├── services/        # AuthService, ProductService, etc.
│   │   ├── guards/          # AuthGuard, SellerGuard
│   │   └── interceptors/    # Auth, Error interceptors
│   ├── features/            # Módulos de funcionalidad
│   │   ├── auth/           # Login, registro
│   │   ├── marketplace/    # Listado de productos, detalle
│   │   ├── dashboard/      # Panel de usuario
│   │   ├── seller/         # Panel de vendedor
│   │   └── profile/        # Perfil de usuario
│   └── shared/             # Componentes compartidos
│       ├── components/     # LoadingSpinner, ImageCarousel
│       └── pipes/          # CurrencyFormat, TimeAgo
└── environments/           # Configuración por ambiente
```

## 🔗 Integración con Backend

El frontend se conecta automáticamente con el backend Go ubicado en `../agro-mas-backend`:

### Endpoints principales:
- `POST /auth/login` - Autenticación
- `GET /products` - Listado de productos
- `POST /products` - Crear producto (sellers)
- `GET /transactions` - Historial de transacciones
- `POST /inquiries` - Crear consulta sobre producto

### Autenticación:
- JWT tokens almacenados en localStorage
- Auto-refresh de tokens con interceptor
- Guards para proteger rutas por roles

## 📱 Features Implementadas

### ✅ Core
- [x] Autenticación JWT con refresh tokens
- [x] Guards de rutas y roles (buyer/seller/admin)
- [x] Interceptors para auth y manejo de errores
- [x] Responsive design con Angular Material

### ✅ Marketplace
- [x] Listado de productos con paginación
- [x] Filtros por categoría, precio, ubicación
- [x] Búsqueda de productos
- [x] Vista detallada de producto
- [x] Sistema de favoritos

### ✅ Gestión de Productos (Sellers)
- [x] CRUD de productos
- [x] Upload de imágenes múltiples
- [x] Geolocalización de productos
- [x] Gestión de stock y precios

### ✅ Transacciones
- [x] Sistema de inquiries/consultas
- [x] Historial de transacciones
- [x] Sistema de reviews y ratings

## 🚀 Deploy a Firebase

### Configuración Automática

El proyecto incluye GitHub Actions para deploy automático:

- **Push a `develop`** → Deploy a Firebase proyecto dev
- **Push a `main`** → Deploy a Firebase proyecto prod
- **Pull Requests** → Preview deploys

### Deploy Manual

```bash
# Build para producción
npm run build:prod

# Deploy a Firebase
firebase deploy
```

## 🧪 Testing

```bash
# Unit tests
npm run test

# E2E tests
npm run e2e

# Linting
npm run lint
```

## 📊 Scripts Disponibles

```bash
npm start              # Desarrollo (puerto 4200)
npm run build          # Build development
npm run build:prod     # Build production
npm run test           # Tests unitarios
npm run lint           # Linting con ESLint
npm run deploy         # Build + Firebase deploy
```

## 🔧 Desarrollo

### Generar Componentes

```bash
# Generar componente
ng generate component features/marketplace/components/product-search

# Generar servicio
ng generate service core/services/notification

# Generar guard
ng generate guard core/guards/admin
```

### Estructura de Commits

```bash
feat: agregar filtro por categoría en marketplace
fix: corregir problema de autenticación en mobile
docs: actualizar README con instrucciones de deploy
style: mejorar responsive design en product cards
```

## 🌐 Variables de Configuración

| Variable | Desarrollo | Producción | Descripción |
|----------|------------|------------|-------------|
| `apiUrl` | `http://localhost:8080/api/v1` | `https://agro-mas-backend-xxx.run.app/api/v1` | URL del backend Go |
| `firebaseConfig.projectId` | `agro-mas-dev` | `agro-mas-prod` | ID proyecto Firebase |

## 📄 Licencia

Este proyecto es parte del ecosistema Agro Mas para conectar el sector agropecuario argentino.

## 🤝 Contribuir

1. Fork el repositorio
2. Crear branch feature (`git checkout -b feature/nueva-funcionalidad`)
3. Commit changes (`git commit -am 'feat: agregar nueva funcionalidad'`)
4. Push branch (`git push origin feature/nueva-funcionalidad`)
5. Crear Pull Request

## 📞 Soporte

Para soporte técnico o consultas sobre la integración con el backend Go, contactar al equipo de desarrollo.