#!/bin/bash

# Script para inicializar o Módulo 4 - IA Jurídica
# Uso: ./scripts/start-ia-module.sh

set -e

echo "🚀 Iniciando Módulo 4 - IA Jurídica do Direito Lux"

# Verificar se Docker está rodando
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker não está rodando. Inicie o Docker primeiro."
    exit 1
fi

# Criar rede se não existir
echo "🌐 Verificando rede direito-lux-network..."
if ! docker network ls | grep -q direito-lux-network; then
    echo "📡 Criando rede direito-lux-network..."
    docker network create direito-lux-network
else
    echo "✅ Rede direito-lux-network já existe"
fi

# Criar arquivo .env se não existir
if [ ! -f "./services/ia-juridica/.env" ]; then
    echo "📝 Criando arquivo .env padrão..."
    cp "./services/ia-juridica/.env.example" "./services/ia-juridica/.env"
    echo "⚠️  IMPORTANTE: Configure suas chaves de API no arquivo .env"
    echo "   - OPENAI_API_KEY (obrigatório para OpenAI)"
    echo "   - GOOGLE_CLOUD_PROJECT (opcional para Vertex AI)"
fi

# Iniciar serviços de IA
echo "🐳 Iniciando serviços do docker-compose.ia.yml..."
docker-compose -f docker-compose.ia.yml up -d

echo "⏳ Aguardando serviços ficarem saudáveis..."
sleep 30

# Verificar status dos serviços
echo "🔍 Verificando status dos serviços..."

# ChromaDB
echo -n "📊 ChromaDB (localhost:8000): "
if curl -s -f http://localhost:8000/api/v1/heartbeat > /dev/null; then
    echo "✅ Online"
else
    echo "❌ Offline"
fi

# Redis
echo -n "⚡ Redis (localhost:6379): "
if docker exec direito-lux-redis redis-cli --no-auth-warning -a "redis123" ping > /dev/null 2>&1; then
    echo "✅ Online"
else
    echo "❌ Offline"
fi

# IA Jurídica Service
echo -n "🧠 IA Jurídica (localhost:9003): "
if curl -s -f http://localhost:9003/health > /dev/null; then
    echo "✅ Online"
else
    echo "❌ Offline - Verificar logs: docker logs direito-lux-ia-juridica"
fi

# Celery Flower
echo -n "🌸 Celery Flower (localhost:5555): "
if curl -s -f http://localhost:5555 > /dev/null; then
    echo "✅ Online"
else
    echo "❌ Offline"
fi

echo ""
echo "📋 URLs disponíveis:"
echo "   🧠 IA Jurídica API: http://localhost:9003"
echo "   📖 Documentação: http://localhost:9003/docs"
echo "   📊 ChromaDB: http://localhost:8000"
echo "   🌸 Celery Monitor: http://localhost:5555"
echo "   ❤️  Health Check: http://localhost:9003/health"

echo ""
echo "🎯 Próximos passos:"
echo "   1. Configure OPENAI_API_KEY no arquivo .env"
echo "   2. Reinicie: docker-compose -f docker-compose.ia.yml restart ia-juridica"
echo "   3. Inicialize a base: python scripts/setup-knowledge-base.py init"
echo "   4. Teste no Postman com: POST http://localhost:9003/api/v1/rag/query"

echo ""
echo "✅ Setup concluído! Serviços rodando em background."