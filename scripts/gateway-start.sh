#!/bin/bash

echo "🚀 Iniciando Módulo 2 - API Gateway, Health, OPA"
echo ""

# Criar rede se não existir
echo "🔧 Criando rede Docker..."
docker network create direito-lux-network 2>/dev/null || {
    echo "ℹ️  Rede direito-lux-network já existe"
}

# Verificar se há conflitos de porta
echo "🔍 Verificando portas..."
ports_to_check=(8002 8003 8181 9090 3000 16686)
ports_in_use=()

for port in "${ports_to_check[@]}"; do
    if lsof -ti:$port >/dev/null 2>&1; then
        ports_in_use+=($port)
    fi
done

if [ ${#ports_in_use[@]} -gt 0 ]; then
    echo "⚠️  Portas em uso detectadas: ${ports_in_use[*]}"
    echo "   Você pode parar os processos ou continuar (alguns serviços podem falhar)"
    read -p "   Continuar mesmo assim? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ Cancelado pelo usuário"
        exit 1
    fi
fi

# Iniciar serviços do gateway
echo "🐳 Iniciando serviços do gateway..."
docker compose -f docker-compose.gateway.yml up -d

echo ""
echo "⏳ Aguardando serviços ficarem prontos..."
sleep 15

# Verificar status dos serviços
echo ""
echo "🔍 Verificando status dos serviços..."

services=(
    "kong-postgres:5432:PostgreSQL Kong"
    "kong-gateway:8003:Kong Admin API"
    "opa-server:8181:Open Policy Agent"
    "prometheus:9090:Prometheus"
    "grafana:3000:Grafana"
    "jaeger:16686:Jaeger UI"
)

all_healthy=true

for service in "${services[@]}"; do
    IFS=':' read -r container port name <<< "$service"
    
    if docker ps --format "table {{.Names}}\t{{.Status}}" | grep -q "$container.*Up"; then
        if nc -z localhost $port 2>/dev/null; then
            echo "✅ $name - http://localhost:$port"
        else
            echo "⚠️  $name - Container rodando mas porta $port não acessível"
            all_healthy=false
        fi
    else
        echo "❌ $name - Container não está rodando"
        all_healthy=false
    fi
done

echo ""

if [ "$all_healthy" = true ]; then
    echo "✅ MÓDULO 2 INICIADO COM SUCESSO!"
else
    echo "⚠️  Alguns serviços podem não estar totalmente prontos ainda"
    echo "   Aguarde mais alguns minutos e verifique novamente"
fi

echo ""
echo "📊 URLs dos Serviços:"
echo "   🚪 Kong Gateway:    http://localhost:8002"
echo "   🔧 Kong Admin:      http://localhost:8003"
echo "   🛡️  OPA:             http://localhost:8181"
echo "   📈 Prometheus:      http://localhost:9090"
echo "   📊 Grafana:         http://localhost:3000 (admin/admin)"
echo "   🔍 Jaeger:          http://localhost:16686"
echo ""
echo "🔧 Para parar o ambiente:"
echo "   ./scripts/gateway-stop.sh"
echo ""
echo "📊 Para ver logs:"
echo "   docker compose -f docker-compose.gateway.yml logs -f"