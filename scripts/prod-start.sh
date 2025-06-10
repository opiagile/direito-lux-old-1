#!/bin/bash

echo "🚀 Iniciando ambiente de PRODUÇÃO/DEMO"
echo "   - Tudo containerizado (API Go + Infraestrutura)"
echo ""

# Para containers de desenvolvimento se estiverem rodando
echo "🔄 Parando ambiente de desenvolvimento..."
docker compose -f docker-compose.dev.yml down 2>/dev/null || true

# Para API Go local se estiver rodando
pkill -f "go run" 2>/dev/null || true

# Para containers da configuração anterior
docker compose -f docker-compose.yml down 2>/dev/null || true

# Build e inicia ambiente de produção
echo "🔧 Construindo e iniciando ambiente de produção..."
docker compose -f docker-compose.prod.yml up --build -d

echo ""
echo "⏳ Aguardando serviços ficarem prontos..."
sleep 15

# Verifica serviços
echo "🔍 Verificando serviços..."

# PostgreSQL
docker exec direito-lux-postgres pg_isready -U keycloak && echo "✅ PostgreSQL pronto" || echo "❌ PostgreSQL com problema"

# Redis
docker exec direito-lux-redis redis-cli ping | grep -q PONG && echo "✅ Redis pronto" || echo "❌ Redis com problema"

# Keycloak
curl -s -f http://localhost:8080 > /dev/null && echo "✅ Keycloak pronto" || echo "⚠️  Keycloak ainda inicializando"

# API Go
curl -s -f http://localhost:9001/health > /dev/null && echo "✅ API Go pronta" || echo "⚠️  API Go ainda inicializando"

# Cria banco da aplicação se não existir
echo "🗄️  Configurando banco de dados..."
docker exec direito-lux-postgres psql -U keycloak -d postgres -c "CREATE DATABASE direito_lux OWNER keycloak;" 2>/dev/null || true

# Setup do realm Keycloak
echo "🔐 Configurando realm Keycloak..."
sleep 5
./scripts/keycloak-realm-setup.sh 2>/dev/null || true

echo ""
echo "✅ AMBIENTE DE PRODUÇÃO/DEMO PRONTO!"
echo ""
echo "📊 Serviços disponíveis:"
echo "   🌐 API Go:      http://localhost:9001"
echo "   🔐 Keycloak:    http://localhost:8080 (admin/admin)" 
echo "   🔧 Nginx:       http://localhost:80"
echo ""
echo "🔧 Para parar o ambiente:"
echo "   ./scripts/prod-stop.sh"
echo ""
echo "📊 Para ver logs:"
echo "   docker compose -f docker-compose.prod.yml logs -f"