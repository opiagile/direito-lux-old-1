#!/bin/bash

# Script rápido para testar billing antes do setup completo

echo "🔍 Testando configuração de billing..."

# Verificar se gcloud está instalado
if ! command -v gcloud &> /dev/null; then
    echo "❌ gcloud CLI não encontrado!"
    exit 1
fi

# Verificar autenticação
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q "@"; then
    echo "❌ Não autenticado no gcloud"
    echo "Execute: gcloud auth login"
    exit 1
fi

echo "✅ gcloud autenticado"

# Listar billing accounts
echo ""
echo "📋 Billing accounts disponíveis:"
gcloud billing accounts list

# Pegar ID da primeira billing account ativa
billing_id=$(gcloud billing accounts list --format="value(name)" --filter="open:true" | head -n1)

if [[ -z "$billing_id" ]]; then
    echo ""
    echo "❌ Nenhuma billing account ativa encontrada!"
    echo ""
    echo "📝 Para resolver:"
    echo "   1. Acesse: https://console.cloud.google.com/billing"
    echo "   2. Crie uma billing account"
    echo "   3. Adicione método de pagamento"
    echo ""
    exit 1
fi

echo ""
echo "✅ Billing account ativa encontrada: $billing_id"
echo ""
echo "🎯 Configuração OK! Você pode executar:"
echo "   ./scripts/configure-github-secrets.sh setup"