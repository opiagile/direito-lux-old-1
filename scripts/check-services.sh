#!/bin/bash

# Script para verificar status de todos os serviços

echo "🔍 Verificando status dos serviços Direito Lux..."
echo "=============================================="
echo ""

# Função para testar endpoint
test_endpoint() {
    local name=$1
    local url=$2
    local expected=$3
    
    if curl -s -f "$url" > /dev/null 2>&1; then
        echo "✅ $name: OK ($url)"
    else
        echo "❌ $name: FALHA ($url)"
    fi
}

# Verificar containers
echo "📦 Docker Containers:"
docker-compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
echo ""

# Testar endpoints
echo "🌐 Endpoints:"
test_endpoint "Keycloak" "http://localhost:8080/realms/master" "200"
test_endpoint "Kong Gateway" "http://localhost:8002" "200"
test_endpoint "Kong Admin" "http://localhost:8003" "200"
test_endpoint "Grafana" "http://localhost:3000" "200"
test_endpoint "Prometheus" "http://localhost:9090" "200"
test_endpoint "Jaeger" "http://localhost:16686" "200"
test_endpoint "OPA" "http://localhost:8181/v1/data" "200"
echo ""

# Verificar logs de containers com problemas
echo "📋 Containers com problemas:"
docker-compose ps --format json | jq -r '. | select(.Status | contains("unhealthy")) | .Name' | while read container; do
    echo "⚠️  $container está unhealthy"
    echo "   Últimas linhas do log:"
    docker logs --tail 5 "$container" 2>&1 | sed 's/^/   /'
    echo ""
done

echo ""
echo "💡 Dicas:"
echo "- Para ver logs completos: docker logs <container-name>"
echo "- Para reiniciar um serviço: docker-compose restart <service-name>"
echo "- Para parar tudo: docker-compose down"
echo "- Para iniciar tudo: docker-compose up -d"