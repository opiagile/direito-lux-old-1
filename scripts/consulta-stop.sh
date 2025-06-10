#!/bin/bash

echo "🛑 Parando Módulo 3 - Consulta Jurídica + Circuit Breaker + ELK"

# Para todos os containers do módulo 3
echo "🐳 Parando containers Docker..."
docker compose -f docker-compose.consulta.yml down

echo ""
echo "✅ Módulo 3 parado!"
echo ""
echo "🚀 Para iniciar novamente:"
echo "   ./scripts/consulta-start.sh"