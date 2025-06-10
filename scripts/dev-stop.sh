#!/bin/bash

echo "🛑 Parando ambiente de DESENVOLVIMENTO"

# Para API Go local se estiver rodando
echo "📱 Parando API Go local..."
pkill -f "go run cmd/demo/simple.go" 2>/dev/null || true
pkill -f "go run cmd/api/main.go" 2>/dev/null || true

# Para containers de desenvolvimento
echo "🐳 Parando containers Docker..."
docker compose -f docker-compose.dev.yml down

echo ""
echo "✅ Ambiente de desenvolvimento parado!"
echo ""
echo "🚀 Para iniciar novamente:"
echo "   ./scripts/dev-start.sh"