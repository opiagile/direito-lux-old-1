#!/bin/bash

echo "🚀 Iniciando ambiente de DESENVOLVIMENTO"
echo "   - Keycloak, PostgreSQL, Redis no Docker"
echo "   - API Go rodará localmente para desenvolvimento rápido"
echo ""

# Para containers existentes da configuração anterior
echo "📦 Parando containers da configuração anterior..."
docker compose -f docker-compose.yml down 2>/dev/null || true

# Inicia infraestrutura para desenvolvimento
echo "🔧 Iniciando infraestrutura (Docker)..."
docker compose -f docker-compose.dev.yml up -d

echo ""
echo "⏳ Aguardando serviços ficarem prontos..."
sleep 10

# Verifica se PostgreSQL está pronto
echo "🔍 Verificando PostgreSQL..."
docker exec direito-lux-postgres-dev pg_isready -U keycloak || {
    echo "❌ PostgreSQL não está pronto"
    exit 1
}

# Verifica se Redis está pronto
echo "🔍 Verificando Redis..."
docker exec direito-lux-redis-dev redis-cli ping | grep -q PONG || {
    echo "❌ Redis não está pronto"
    exit 1
}

# Verifica se Keycloak está pronto
echo "🔍 Verificando Keycloak..."
curl -s -f http://localhost:8080 > /dev/null || {
    echo "⚠️  Keycloak ainda não está pronto (pode levar alguns minutos)"
}

# Cria banco da aplicação se não existir
echo "🗄️  Criando banco de dados da aplicação..."
docker exec direito-lux-postgres-dev psql -U keycloak -d postgres -c "CREATE DATABASE direito_lux OWNER keycloak;" 2>/dev/null || {
    echo "ℹ️  Banco direito_lux já existe"
}

# Setup do realm Keycloak (se necessário)
echo "🔐 Configurando realm Keycloak..."
sleep 5
./scripts/keycloak-realm-setup.sh 2>/dev/null || {
    echo "ℹ️  Realm já configurado ou Keycloak ainda inicializando"
}

echo ""
echo "✅ AMBIENTE DE DESENVOLVIMENTO PRONTO!"
echo ""
echo "📊 Serviços rodando:"
echo "   🔐 Keycloak:    http://localhost:8080 (admin/admin)"
echo "   🗄️  PostgreSQL:  localhost:5432 (keycloak/keycloak)"
echo "   📦 Redis:       localhost:6379"
echo ""
echo "🚀 Para iniciar a API Go localmente:"
echo "   go run cmd/demo/simple.go"
echo ""
echo "🔧 Para parar o ambiente:"
echo "   ./scripts/dev-stop.sh"