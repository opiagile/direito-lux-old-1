#!/bin/bash

echo "🔍 MÓDULO 2 - STATUS DOS SERVIÇOS"
echo ""

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
    echo "✅ MÓDULO 2 FUNCIONANDO PERFEITAMENTE!"
else
    echo "⚠️  Alguns serviços podem não estar totalmente prontos ainda"
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
echo "🧪 Testando Kong Gateway:"
if curl -f http://localhost:8002/health >/dev/null 2>&1; then
    echo "✅ Kong Gateway está roteando corretamente"
else
    echo "⚠️  Kong Gateway pode precisar de configuração"
    echo "   Execute: ./scripts/kong-setup.sh"
fi

echo ""
echo "🔧 Comandos úteis:"
echo "   ./scripts/gateway-stop.sh   - Parar ambiente"
echo "   ./scripts/kong-setup.sh     - Configurar rotas Kong"