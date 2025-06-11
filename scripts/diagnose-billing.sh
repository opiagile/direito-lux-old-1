#!/bin/bash

# Script de diagnóstico para problemas de billing

echo "🔍 Diagnóstico completo de billing..."

project="direito-lux-dev"
billing_id="01B2F9-AD5BB4-BE339E"

echo ""
echo "📊 Informações do projeto: $project"
echo "========================================="

# Verificar se projeto existe
if gcloud projects describe $project >/dev/null 2>&1; then
    echo "✅ Projeto existe"
    
    # Verificar billing atual
    echo ""
    echo "💳 Status atual do billing:"
    current_billing=$(gcloud billing projects describe $project --format="value(billingAccountName)" 2>/dev/null)
    if [[ -n "$current_billing" ]]; then
        echo "✅ Billing já configurado: $current_billing"
        echo "🎯 Não precisa configurar billing novamente!"
    else
        echo "❌ Billing não configurado"
        
        # Tentar vincular
        echo ""
        echo "🔧 Tentando vincular billing account..."
        echo "Comando: gcloud billing projects link $project --billing-account=$billing_id"
        
        if gcloud billing projects link $project --billing-account="$billing_id" --verbosity=debug; then
            echo "✅ Billing configurado com sucesso!"
        else
            echo "❌ Falha ao configurar billing"
            echo ""
            echo "🔍 Possíveis causas:"
            echo "   1. Permissões insuficientes"
            echo "   2. Billing account inválida"
            echo "   3. Projeto em estado inconsistente"
            echo ""
            echo "💡 Soluções:"
            echo "   1. Verificar permissões: gcloud projects get-iam-policy $project"
            echo "   2. Configurar manualmente no console: https://console.cloud.google.com/billing"
        fi
    fi
    
    # Verificar permissões
    echo ""
    echo "🔐 Verificando permissões:"
    user_email=$(gcloud auth list --filter=status:ACTIVE --format="value(account)")
    echo "Usuário atual: $user_email"
    
    # Verificar se é owner/editor do projeto
    iam_policy=$(gcloud projects get-iam-policy $project --format="value(bindings.members)" --flatten="bindings[].members" --filter="bindings.members:$user_email" 2>/dev/null)
    if [[ -n "$iam_policy" ]]; then
        echo "✅ Usuário tem permissões no projeto"
    else
        echo "❌ Usuário pode não ter permissões suficientes"
    fi
    
else
    echo "❌ Projeto não existe"
fi

echo ""
echo "📋 Informações da billing account:"
echo "=================================="
gcloud billing accounts describe $billing_id 2>/dev/null || echo "❌ Billing account inválida"

echo ""
echo "🎯 Próximos passos recomendados:"
echo "1. Se billing já está configurado, pule para próximo projeto"
echo "2. Se não está configurado, configure manualmente no console"
echo "3. Ou delete e recrie o projeto: gcloud projects delete $project"