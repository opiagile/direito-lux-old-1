#!/bin/bash

echo "🧪 TESTANDO MÓDULO 3 - API DE CONSULTA JURÍDICA"
echo ""

BASE_URL="http://localhost:9002"

echo "1. 💚 Testando Health Check..."
curl -s "$BASE_URL/health" | jq . || echo "❌ Health check failed"
echo ""

echo "2. 📈 Testando Circuit Breaker Status..."
curl -s "$BASE_URL/api/v1/circuit-breaker/status" | jq . || echo "❌ Circuit breaker status failed"
echo ""

echo "3. ⚖️  Testando Consulta de Processo..."
curl -s -X POST "$BASE_URL/api/v1/consultas/processos" \
  -H "Content-Type: application/json" \
  -d '{
    "numero_processo": "1234567-89.2023.1.23.4567",
    "tribunal": "TJSP"
  }' | jq . || echo "❌ Consulta de processo failed"
echo ""

echo "4. 📚 Testando Consulta de Legislação..."
curl -s -X POST "$BASE_URL/api/v1/consultas/legislacao" \
  -H "Content-Type: application/json" \
  -d '{
    "tema": "direito civil",
    "jurisdicao": "federal"
  }' | jq . || echo "❌ Consulta de legislação failed"
echo ""

echo "5. 📖 Testando Consulta de Jurisprudência..."
curl -s -X POST "$BASE_URL/api/v1/consultas/jurisprudencia" \
  -H "Content-Type: application/json" \
  -d '{
    "tema": "responsabilidade civil",
    "tribunal": "STJ"
  }' | jq . || echo "❌ Consulta de jurisprudência failed"
echo ""

echo "6. ⚡ Testando Circuit Breaker (processo que falha)..."
curl -s -X POST "$BASE_URL/api/v1/consultas/processos" \
  -H "Content-Type: application/json" \
  -d '{
    "numero_processo": "0000000-00.0000.0.00.0000",
    "tribunal": "FAIL"
  }' | jq . || echo "❌ Teste de falha do circuit breaker failed"
echo ""

echo "7. 📈 Verificando status do Circuit Breaker após falha..."
curl -s "$BASE_URL/api/v1/circuit-breaker/status" | jq . || echo "❌ Circuit breaker status after failure failed"
echo ""

echo "✅ Testes concluídos!"
echo ""
echo "📊 Para ver logs em tempo real:"
echo "   docker logs -f consulta-service"
echo ""
echo "🔍 Para ver métricas no Kibana:"
echo "   http://localhost:5601"