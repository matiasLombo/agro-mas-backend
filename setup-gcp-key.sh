#!/bin/bash

# Script para configurar correctamente el Service Account Key de Google Cloud
echo "🔧 Configurando Google Cloud Service Account Key..."

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Este script te ayudará a configurar correctamente el Service Account Key para GitHub Actions${NC}"
echo ""

# Verificar si gcloud está instalado
if ! command -v gcloud &> /dev/null; then
    echo -e "${RED}❌ gcloud CLI no está instalado. Instálalo desde: https://cloud.google.com/sdk/docs/install${NC}"
    exit 1
fi

echo -e "${GREEN}✅ gcloud CLI encontrado${NC}"

# Verificar autenticación
echo "Verificando autenticación..."
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | head -n1 > /dev/null; then
    echo -e "${RED}❌ No estás autenticado en gcloud. Ejecuta: gcloud auth login${NC}"
    exit 1
fi

CURRENT_USER=$(gcloud auth list --filter=status:ACTIVE --format="value(account)" | head -n1)
echo -e "${GREEN}✅ Autenticado como: $CURRENT_USER${NC}"

# Obtener proyecto actual
CURRENT_PROJECT=$(gcloud config get-value project 2>/dev/null)
if [ -z "$CURRENT_PROJECT" ]; then
    echo -e "${YELLOW}⚠️  No hay proyecto configurado. Lista de proyectos disponibles:${NC}"
    gcloud projects list --format="table(projectId,name)"
    echo ""
    read -p "Ingresa el PROJECT_ID para desarrollo (ej: agro-mas-develop): " PROJECT_ID
    gcloud config set project $PROJECT_ID
else
    echo -e "${GREEN}✅ Proyecto actual: $CURRENT_PROJECT${NC}"
    PROJECT_ID=$CURRENT_PROJECT
fi

# Verificar si el proyecto existe
if ! gcloud projects describe $PROJECT_ID &> /dev/null; then
    echo -e "${RED}❌ El proyecto $PROJECT_ID no existe o no tienes permisos${NC}"
    exit 1
fi

# Crear service account si no existe
SA_NAME="github-actions"
SA_EMAIL="$SA_NAME@$PROJECT_ID.iam.gserviceaccount.com"

echo ""
echo "Creando/verificando Service Account..."

if gcloud iam service-accounts describe $SA_EMAIL &> /dev/null; then
    echo -e "${GREEN}✅ Service Account ya existe: $SA_EMAIL${NC}"
else
    echo "Creando Service Account..."
    gcloud iam service-accounts create $SA_NAME \
        --display-name="GitHub Actions" \
        --description="Service Account para GitHub Actions CI/CD"
    echo -e "${GREEN}✅ Service Account creado: $SA_EMAIL${NC}"
fi

# Asignar roles necesarios
echo ""
echo "Asignando roles necesarios..."

ROLES=(
    "roles/run.admin"
    "roles/storage.admin"
    "roles/cloudsql.admin"
    "roles/secretmanager.admin"
    "roles/cloudbuild.builds.editor"
    "roles/iam.serviceAccountUser"
    "roles/compute.admin"
    "roles/serviceusage.serviceUsageAdmin"
)

for ROLE in "${ROLES[@]}"; do
    echo "Asignando rol: $ROLE"
    gcloud projects add-iam-policy-binding $PROJECT_ID \
        --member="serviceAccount:$SA_EMAIL" \
        --role="$ROLE" \
        --quiet
done

echo -e "${GREEN}✅ Roles asignados correctamente${NC}"

# Crear clave del service account
echo ""
echo "Creando clave del Service Account..."

KEY_FILE="$SA_NAME-key.json"
gcloud iam service-accounts keys create $KEY_FILE \
    --iam-account=$SA_EMAIL

if [ -f "$KEY_FILE" ]; then
    echo -e "${GREEN}✅ Clave creada: $KEY_FILE${NC}"
    
    # Comprimir el JSON en una línea (importante para GitHub Secrets)
    COMPRESSED_KEY=$(cat $KEY_FILE | jq -c .)
    
    echo ""
    echo -e "${YELLOW}📋 IMPORTANTE: Copia el siguiente contenido y úsalo como valor del secret GCP_SA_KEY en GitHub:${NC}"
    echo ""
    echo "----------------------------------------"
    echo "$COMPRESSED_KEY"
    echo "----------------------------------------"
    echo ""
    
    # Instrucciones para GitHub
    echo -e "${YELLOW}📝 Para configurar el secret en GitHub:${NC}"
    echo "1. Ve a tu repositorio en GitHub"
    echo "2. Settings > Secrets and variables > Actions"
    echo "3. Actualiza el secret 'GCP_SA_KEY' con el JSON de arriba"
    echo ""
    
    # Limpiar archivo temporal
    rm $KEY_FILE
    echo -e "${GREEN}✅ Archivo temporal limpiado${NC}"
    
    echo ""
    echo -e "${GREEN}🎉 ¡Configuración completada! Ahora puedes hacer push para probar el deploy.${NC}"
else
    echo -e "${RED}❌ Error al crear la clave del service account${NC}"
    exit 1
fi