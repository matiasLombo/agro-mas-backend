# 🚀 Guía Completa de Deployment - Agro Mas Backend

Esta es la documentación completa de todo lo que se tuvo que hacer para configurar el CI/CD de deployment automático en Google Cloud Platform.

## 📋 Resumen del Sistema

**Arquitectura Multi-Ambiente:**
- **Desarrollo**: Rama `develop` → Proyecto `agro-mas-develop`
- **Producción**: Rama `master` → Proyecto `agro-mas-production`

**URLs Finales:**
- 🛠️ **Desarrollo**: https://agro-mas-backend-dev-62ylvs6mya-uc.a.run.app
- 🚀 **Producción**: https://agro-mas-backend-27kvnrkepa-uc.a.run.app

## 🏗️ Infraestructura Configurada

### Proyectos de Google Cloud
1. **agro-mas-develop** (ID: 710155878514)
   - Cloud Run: `agro-mas-backend-dev`
   - Cloud SQL: `agro-mas-db-dev` (PostgreSQL)
   - Base de datos: `agro_mas_dev`
   - Secrets: `agro-mas-jwt-secret-dev`, `agro-mas-db-password-dev`

2. **agro-mas-production** (ID: 51677535367)
   - Cloud Run: `agro-mas-backend`
   - Cloud SQL: `agro-mas-db-prod` (PostgreSQL)
   - Base de datos: `agro_mas_prod`
   - Secrets: `agro-mas-jwt-secret-prod`, `agro-mas-db-password-prod`

## 🔐 Service Accounts Configurados

### 1. Service Account de Desarrollo
- **Email**: `github-actions@agro-mas-develop.iam.gserviceaccount.com`
- **Proyecto**: agro-mas-develop
- **GitHub Secret**: `GCP_SA_KEY`

### 2. Service Account de Producción
- **Email**: `github-actions@agro-mas-production.iam.gserviceaccount.com`
- **Proyecto**: agro-mas-production
- **GitHub Secret**: `GCP_SA_KEY_PROD`

### Roles Asignados a Ambos Service Accounts:
```bash
roles/run.admin
roles/storage.admin
roles/cloudsql.admin
roles/secretmanager.admin
roles/cloudbuild.builds.editor
roles/iam.serviceAccountUser
roles/serviceusage.serviceUsageAdmin
roles/artifactregistry.admin
roles/resourcemanager.projectIamAdmin
```

## 🛠️ Problemas Encontrados y Soluciones

### 1. **Service Account Key con Formato Incorrecto**
**Problema**: Los JSON keys de service account contenían saltos de línea que rompían los GitHub Secrets.

**Solución**:
```bash
# Generar y comprimir en una línea
python -c "import json; print(json.dumps(json.load(open('key.json'))))"
```

### 2. **Service Account de Desarrollo en Producción**
**Problema**: El workflow usaba el mismo service account para ambos ambientes.

**Solución**: Creamos service accounts separados y modificamos el workflow:
```yaml
env:
  GOOGLE_APPLICATION_CREDENTIALS_JSON: ${{ needs.setup.outputs.environment == 'production' && secrets.GCP_SA_KEY_PROD || secrets.GCP_SA_KEY }}
```

### 3. **APIs No Habilitadas**
**Problema**: Faltaban APIs cruciales en el proyecto de producción.

**Solución**:
```bash
gcloud services enable cloudresourcemanager.googleapis.com
gcloud services enable artifactregistry.googleapis.com
gcloud services enable cloudsql.googleapis.com
gcloud services enable secretmanager.googleapis.com
gcloud services enable run.googleapis.com
gcloud services enable cloudbuild.googleapis.com
```

### 4. **Límite de Conexiones de Base de Datos**
**Problema**: `db-f1-micro` solo permite 25 conexiones, pero el código usaba 25 conexiones por instancia.

**Solución**: Reducir el pool de conexiones en `internal/storage/database.go`:
```go
// Antes
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(25)

// Después
db.SetMaxOpenConns(10)
db.SetMaxIdleConns(5)
```

### 5. **Secrets Corruptos**
**Problema**: Los secrets de base de datos contenían caracteres CRLF que rompían las URLs de conexión.

**Solución**:
```bash
# Recrear secrets sin saltos de línea
printf "password_here" | gcloud secrets create secret-name --data-file=-
```

### 6. **Permisos de Artifact Registry**
**Problema**: El service account de producción no podía subir imágenes Docker.

**Solución**:
```bash
gcloud projects add-iam-policy-binding agro-mas-production \
  --member="serviceAccount:github-actions@agro-mas-production.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.admin"
```

## 📁 Archivos de Configuración Clave

### `.github/workflows/deploy.yml`
- Workflow principal que maneja ambos ambientes
- Usa conditional logic para seleccionar service accounts
- Ejecuta: test → infrastructure → backend → integration-test

### `internal/storage/database.go`
- Configuración del pool de conexiones optimizada
- Límites ajustados para `db-f1-micro`

### Scripts de Deployment
- `setup-gcp-key.sh`: Script para generar service account keys correctamente

## 🚀 Proceso de Deployment

### Desarrollo (Branch `develop`)
1. Push a `develop` → Trigger automático
2. Service account: `github-actions@agro-mas-develop.iam.gserviceaccount.com`
3. Deploy a: `agro-mas-develop` project
4. URL: https://agro-mas-backend-dev-62ylvs6mya-uc.a.run.app

### Producción (Branch `master`)
1. Push/merge a `master` → Trigger automático
2. Service account: `github-actions@agro-mas-production.iam.gserviceaccount.com`
3. Deploy a: `agro-mas-production` project
4. URL: https://agro-mas-backend-27kvnrkepa-uc.a.run.app

## 🧪 Endpoints de Verificación

### Health Check
- **Desarrollo**: `/health`
- **Producción**: `/health`
- **Respuesta**: Status de la aplicación y base de datos

### Environment Test
- **Desarrollo**: `/env` → Mensaje azul "🛠️ DEVELOPMENT environment"
- **Producción**: `/env` → Mensaje rojo "🚀 PRODUCTION environment"

## 📋 GitHub Secrets Configurados

```
GCP_SA_KEY      - Service account para desarrollo
GCP_SA_KEY_PROD - Service account para producción
```

## ⚡ Comandos Útiles para Troubleshooting

### Ver logs de Cloud Run
```bash
gcloud logging read "resource.type=\"cloud_run_revision\"" --limit=20
```

### Verificar service account permissions
```bash
gcloud projects get-iam-policy PROJECT_ID
```

### Verificar APIs habilitadas
```bash
gcloud services list --enabled
```

### Verificar secrets
```bash
gcloud secrets list
```

## 🎯 Lecciones Aprendidas

1. **Separar service accounts por ambiente** es crucial para seguridad
2. **Comprimir JSON keys** evita problemas con saltos de línea
3. **Configurar pools de conexión** según límites de la base de datos
4. **Habilitar todas las APIs** antes del primer deployment
5. **Usar environments separados** en GitHub Actions para mejor control

## 🔮 Próximos Pasos

- [ ] Configurar monitoring y alertas
- [ ] Implementar rollback automático
- [ ] Agregar tests de performance
- [ ] Configurar CDN para assets estáticos
- [ ] Implementar blue-green deployments

---
**Creado**: 2025-08-15  
**Última actualización**: Después de resolver todos los problemas de deployment  
**Estado**: ✅ Funcionando correctamente en ambos ambientes