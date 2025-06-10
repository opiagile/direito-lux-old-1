#!/bin/bash

echo "🔍 MÓDULO 3 - STATUS DOS SERVIÇOS"
echo ""

services=(
    "elasticsearch:9200:Elasticsearch"
    "kibana:5601:Kibana" 
    "logstash:9600:Logstash"
    "consulta-service:9002:Consulta Service"
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
    echo "✅ MÓDULO 3 FUNCIONANDO PERFEITAMENTE!"
else
    echo "⚠️  Alguns serviços podem não estar totalmente prontos ainda"
fi

# Testar API de consulta
echo ""
echo "🧪 Testando API de Consulta:"
if curl -f http://localhost:9002/health >/dev/null 2>&1; then
    echo "✅ API de Consulta funcionando"
    
    # Testar circuit breaker status
    if curl -f http://localhost:9002/api/v1/circuit-breaker/status >/dev/null 2>&1; then
        echo "✅ Circuit Breaker funcionando"
    else
        echo "⚠️  Circuit Breaker pode não estar acessível"
    fi
else
    echo "❌ API de Consulta não está respondendo"
fi

# Testar Elasticsearch
echo ""
echo "🔍 Testando Elasticsearch:"
if curl -f http://localhost:9200/_cluster/health >/dev/null 2>&1; then
    echo "✅ Elasticsearch funcionando"
else
    echo "❌ Elasticsearch não está respondendo"
fi

echo ""
echo "📊 URLs dos Serviços:"
echo "   🔍 Elasticsearch:   http://localhost:9200"
echo "   📊 Kibana:          http://localhost:5601"
echo "   📋 Logstash:        http://localhost:9600"
echo "   ⚖️  Consulta API:    http://localhost:9002"
echo ""
echo "🔧 Comandos úteis:"
echo "   ./scripts/consulta-stop.sh    - Parar ambiente"
echo "   ./scripts/consulta-test.sh    - Testar consultas"