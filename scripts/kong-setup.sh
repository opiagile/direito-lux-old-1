#!/bin/bash

echo "🔧 Configurando Kong Gateway..."

# URL do Kong Admin API
KONG_ADMIN="http://localhost:8003"

# Verificar se Kong está rodando
if ! curl -f "$KONG_ADMIN/status" >/dev/null 2>&1; then
    echo "❌ Kong Admin API não está acessível em $KONG_ADMIN"
    exit 1
fi

echo "✅ Kong Admin API acessível"

# Criar serviço demo-api se não existir
SERVICE_EXISTS=$(curl -s "$KONG_ADMIN/services/demo-api" | jq -r '.id // empty' 2>/dev/null)

if [ -z "$SERVICE_EXISTS" ]; then
    echo "🔧 Criando serviço demo-api..."
    curl -X POST "$KONG_ADMIN/services" \
      -H "Content-Type: application/json" \
      -d '{
        "name": "demo-api",
        "url": "http://host.docker.internal:9001",
        "tags": ["demo", "direito-lux"]
      }' >/dev/null 2>&1
    echo "✅ Serviço demo-api criado"
else
    echo "ℹ️  Serviço demo-api já existe"
fi

# Criar rotas
ROUTES=(
    '{"name":"health","paths":["/health"],"strip_path":false,"tags":["health"]}'
    '{"name":"root","paths":["/"],"strip_path":false,"tags":["demo"]}'
    '{"name":"api-v1","paths":["/api/v1"],"strip_path":false,"tags":["api"]}'
)

for route_config in "${ROUTES[@]}"; do
    route_name=$(echo "$route_config" | jq -r '.name')
    
    # Verificar se rota já existe
    ROUTE_EXISTS=$(curl -s "$KONG_ADMIN/routes" | jq -r ".data[] | select(.name == \"$route_name\") | .id" 2>/dev/null)
    
    if [ -z "$ROUTE_EXISTS" ]; then
        echo "🔧 Criando rota $route_name..."
        curl -X POST "$KONG_ADMIN/services/demo-api/routes" \
          -H "Content-Type: application/json" \
          -d "$route_config" >/dev/null 2>&1
        echo "✅ Rota $route_name criada"
    else
        echo "ℹ️  Rota $route_name já existe"
    fi
done

# Adicionar plugin CORS global se não existir
CORS_EXISTS=$(curl -s "$KONG_ADMIN/plugins" | jq -r '.data[] | select(.name == "cors") | .id' 2>/dev/null)

if [ -z "$CORS_EXISTS" ]; then
    echo "🔧 Adicionando plugin CORS..."
    curl -X POST "$KONG_ADMIN/plugins" \
      -H "Content-Type: application/json" \
      -d '{
        "name": "cors",
        "config": {
          "origins": ["*"],
          "methods": ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
          "headers": ["Accept", "Content-Type", "Authorization", "X-Request-ID"],
          "credentials": true
        }
      }' >/dev/null 2>&1
    echo "✅ Plugin CORS adicionado"
else
    echo "ℹ️  Plugin CORS já existe"
fi

echo ""
echo "✅ Kong Gateway configurado com sucesso!"
echo ""
echo "📊 URLs disponíveis:"
echo "   🚪 Kong Gateway:    http://localhost:8002"
echo "   🔧 Kong Admin:      http://localhost:8003"
echo "   💚 Health Check:    http://localhost:8002/health"
echo "   🏠 Demo Page:       http://localhost:8002/"
echo ""