#!/bin/bash

echo "🚀 Iniciando Módulo 3 - Consulta Jurídica + Circuit Breaker + ELK"
echo ""

# Criar rede se não existir
echo "🔧 Criando rede Docker..."
docker network create direito-lux-network 2>/dev/null || {
    echo "ℹ️  Rede direito-lux-network já existe"
}

# Verificar se há conflitos de porta
echo "🔍 Verificando portas..."
ports_to_check=(9200 5601 5044 9002)
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

# Iniciar serviços do ELK + Consulta
echo "🐳 Iniciando serviços do Módulo 3..."
docker compose -f docker-compose.consulta.yml up -d

echo ""
echo "⏳ Aguardando serviços ficarem prontos..."
sleep 30

# Verificar status dos serviços
echo ""
echo "🔍 Verificando status dos serviços..."

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
    echo "✅ MÓDULO 3 INICIADO COM SUCESSO!"
else
    echo "⚠️  Alguns serviços podem não estar totalmente prontos ainda"
    echo "   Aguarde mais alguns minutos e verifique novamente"
fi

echo ""
echo "📊 URLs dos Serviços:"
echo "   🔍 Elasticsearch:   http://localhost:9200"
echo "   📊 Kibana:          http://localhost:5601"
echo "   📋 Logstash:        http://localhost:9600"
echo "   ⚖️  Consulta API:    http://localhost:9002"
echo ""
echo "🧪 Endpoints de Teste:"
echo "   💚 Health Check:    http://localhost:9002/health"
echo "   📈 Circuit Breaker: http://localhost:9002/api/v1/circuit-breaker/status"
echo ""
echo "🔧 Para parar o ambiente:"
echo "   ./scripts/consulta-stop.sh"
echo ""
echo "📊 Para ver logs:"
echo "   docker compose -f docker-compose.consulta.yml logs -f"