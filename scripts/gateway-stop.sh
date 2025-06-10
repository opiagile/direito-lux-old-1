#!/bin/bash

echo "🛑 Parando Módulo 2 - API Gateway, Health, OPA"

# Para todos os containers do gateway
echo "🐳 Parando containers Docker..."
docker compose -f docker-compose.gateway.yml down

echo ""
echo "✅ Módulo 2 parado!"
echo ""
echo "🚀 Para iniciar novamente:"
echo "   ./scripts/gateway-start.sh"