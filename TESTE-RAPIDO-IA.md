# 🧠 Teste Rápido - Módulo 4 IA Jurídica

## ⚡ Setup em 3 Passos

### 1. 🚀 Iniciar Serviços
```bash
# Executar script de inicialização
./scripts/start-ia-module.sh
```

### 2. 🔑 Configurar Chave da OpenAI
```bash
# Editar arquivo de configuração
nano services/ia-juridica/.env

# Adicionar sua chave OpenAI
DIREITO_LUX_IA_OPENAI_API_KEY=sk-sua-chave-aqui

# Reiniciar serviço
docker-compose -f docker-compose.ia.yml restart ia-juridica
```

### 3. 📚 Inicializar Base de Conhecimento
```bash
# Carregar documentos jurídicos iniciais
cd scripts
python setup-knowledge-base.py init
```

## 🎯 Testes no Postman

### 1. Importar Collection
- Abrir Postman
- Import → File → `postman/Direito-Lux-IA-Module.postman_collection.json`

### 2. Testar Health Check
```
GET http://localhost:9003/health
```
**Resposta esperada:** Status 200 com informações dos serviços

### 3. 🧠 Teste Principal - Consulta Jurídica
```
POST http://localhost:9003/api/v1/rag/query
Content-Type: application/json

{
  "question": "O que são direitos fundamentais e onde estão previstos na Constituição?",
  "query_type": "legislacao",
  "evaluate_response": true
}
```

**Resposta esperada:**
```json
{
  "query_id": "query_1733857200",
  "timestamp": "2024-12-10T20:00:00.000Z",
  "question": "O que são direitos fundamentais...",
  "answer": "Direitos fundamentais são direitos básicos...",
  "query_type": "legislacao",
  "sources": [
    {
      "title": "Constituição Federal - Art. 5º",
      "similarity_score": 0.85,
      "law_number": "Constituição Federal de 1988"
    }
  ],
  "processing_time": 2.5,
  "retrieved_docs_count": 1
}
```

## 🔍 URLs para Teste

| Serviço | URL | Descrição |
|---------|-----|-----------|
| 🧠 IA API | http://localhost:9003 | API principal de IA |
| 📖 Docs | http://localhost:9003/docs | Documentação automática |
| ❤️ Health | http://localhost:9003/health | Status dos serviços |
| 📊 ChromaDB | http://localhost:8000 | Vector database |
| 🌸 Celery | http://localhost:5555 | Monitor de tasks |

## 🧪 Exemplos de Teste

### Consulta sobre Responsabilidade Civil
```json
{
  "question": "Quais são os elementos da responsabilidade civil?",
  "query_type": "legislacao"
}
```

### Consulta sobre Justa Causa
```json
{
  "question": "Quais são as hipóteses de justa causa trabalhista?",
  "query_type": "legislacao",
  "context_filters": {"source_type": "clt"}
}
```

### Busca Semântica
```json
{
  "query": "direitos fundamentais",
  "top_k": 3,
  "min_similarity": 0.5
}
```

## 🛠️ Troubleshooting

### ❌ Erro 503 "Service not available"
```bash
# Verificar logs
docker logs direito-lux-ia-juridica

# Verificar se ChromaDB está rodando
curl http://localhost:8000/api/v1/heartbeat
```

### ❌ Erro OpenAI API
- Verificar se `OPENAI_API_KEY` está configurado corretamente
- Verificar saldo da conta OpenAI

### ❌ Base de conhecimento vazia
```bash
# Verificar status
curl http://localhost:9003/api/v1/knowledge/stats

# Reinicializar se necessário
python scripts/setup-knowledge-base.py init
```

## 📈 Métricas Disponíveis

- **Faithfulness**: Fidelidade ao contexto
- **Answer Relevancy**: Relevância da resposta
- **Context Precision**: Precisão do contexto
- **Context Recall**: Cobertura do contexto
- **Answer Similarity**: Similaridade semântica
- **Answer Correctness**: Correção factual

## 🎉 Próximos Passos

1. ✅ Testar todas as consultas da collection
2. 📊 Verificar métricas de avaliação
3. 📚 Adicionar mais documentos jurídicos
4. 🔄 Testar consultas em lote
5. 🌸 Monitorar tasks no Celery Flower