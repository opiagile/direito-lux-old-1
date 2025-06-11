#!/bin/bash

# Script para instalar ferramentas necessárias no macOS

echo "🔧 Instalando ferramentas necessárias..."

# Verificar se Homebrew está instalado
if ! command -v brew &> /dev/null; then
    echo "❌ Homebrew não encontrado. Instale primeiro: https://brew.sh"
    exit 1
fi

echo "✅ Homebrew encontrado"

# Instalar Helm
if ! command -v helm &> /dev/null; then
    echo "📦 Instalando Helm..."
    brew install helm
else
    echo "✅ Helm já instalado"
fi

# Instalar ArgoCD CLI
if ! command -v argocd &> /dev/null; then
    echo "📦 Instalando ArgoCD CLI..."
    brew install argocd
else
    echo "✅ ArgoCD CLI já instalado"
fi

# Verificar instalações
echo ""
echo "🔍 Verificando instalações:"
echo "Docker: $(docker --version)"
echo "Kubectl: $(kubectl version --client --short 2>/dev/null)"
echo "Terraform: $(terraform --version | head -1)"
echo "gcloud: $(gcloud --version | head -1)"
echo "Helm: $(helm version --short)"
echo "ArgoCD: $(argocd version --client --short)"

echo ""
echo "✅ Todas as ferramentas instaladas!"