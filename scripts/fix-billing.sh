#!/bin/bash

# Script para configurar billing em projetos específicos

set -e

if [[ $# -ne 1 ]]; then
    echo "Uso: $0 <project-name>"
    echo ""
    echo "Exemplos:"
    echo "  $0 direito-lux-staging"
    echo "  $0 direito-lux-prod"
    exit 1
fi

project=$1
billing_id="01B2F9-AD5BB4-BE339E"

echo "🔧 Configurando billing para: $project"

# Verificar se projeto existe
if ! gcloud projects describe $project >/dev/null 2>&1; then
    echo "❌ Projeto $project não existe"
    echo "Criando projeto..."
    gcloud projects create $project --name="Direito Lux - ${project##*-}"
fi

# Configurar billing
echo "💳 Configurando billing account..."
if gcloud billing projects link $project --billing-account="$billing_id"; then
    echo "✅ Billing configurado com sucesso!"
else
    echo "❌ Falha ao configurar billing"
    echo ""
    echo "📝 Configure manualmente:"
    echo "   1. Acesse: https://console.cloud.google.com/billing"
    echo "   2. Vá em 'Manage billing accounts'"
    echo "   3. Selecione 'Minha conta de faturamento'"
    echo "   4. Vá em 'MY PROJECTS'"
    echo "   5. Clique em 'LINK A PROJECT'"
    echo "   6. Selecione: $project"
    exit 1
fi

# Habilitar APIs
echo "🚀 Habilitando APIs..."
gcloud services enable --project=$project \
    container.googleapis.com \
    sqladmin.googleapis.com \
    redis.googleapis.com \
    secretmanager.googleapis.com \
    monitoring.googleapis.com \
    logging.googleapis.com \
    cloudresourcemanager.googleapis.com \
    iam.googleapis.com \
    compute.googleapis.com \
    artifactregistry.googleapis.com \
    containerregistry.googleapis.com \
    dns.googleapis.com

echo "✅ Projeto $project configurado com sucesso!"