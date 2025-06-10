#!/bin/bash

echo "🛑 Parando ambiente de PRODUÇÃO/DEMO"

# Para todos os containers de produção
echo "🐳 Parando containers Docker..."
docker compose -f docker-compose.prod.yml down

echo ""
echo "✅ Ambiente de produção parado!"
echo ""
echo "🚀 Para iniciar novamente:"
echo "   ./scripts/prod-start.sh"
echo ""
echo "🔄 Para voltar ao desenvolvimento:"
echo "   ./scripts/dev-start.sh"