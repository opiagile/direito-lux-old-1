# 🧠 Direito Lux - Serviço de IA Jurídica

## ⚠️ IMPORTANTE: Configuração do .env

Antes de iniciar o serviço, você DEVE criar o arquivo `.env` baseado no `.env.example`:

```bash
cp .env.example .env
```

### 🔑 Configurar a chave OpenAI:

Edite o arquivo `.env` e adicione sua chave OpenAI:

```bash
# IMPORTANTE: Substitua pela sua chave real
DIREITO_LUX_IA_OPENAI_API_KEY=sk-proj-sua-chave-openai-aqui
OPENAI_API_KEY=sk-proj-sua-chave-openai-aqui
```

### 📋 Configuração completa do .env (para referência):

```env
# Direito Lux - IA Jurídica Environment Configuration

# Application Settings
DIREITO_LUX_IA_ENVIRONMENT=development
DIREITO_LUX_IA_PORT=9003
DIREITO_LUX_IA_WORKERS=4
DIREITO_LUX_IA_LOG_LEVEL=INFO

# Security
DIREITO_LUX_IA_SECRET_KEY=direito-lux-super-secret-key-2024

# Vector Database (ChromaDB)
DIREITO_LUX_IA_CHROMA_HOST=localhost
DIREITO_LUX_IA_CHROMA_PORT=8000
DIREITO_LUX_IA_CHROMA_COLLECTION_NAME=direito_lux_legal_docs

# LLM Configuration - OpenAI
DIREITO_LUX_IA_LLM_PROVIDER=openai
DIREITO_LUX_IA_OPENAI_API_KEY=sk-proj-sua-chave-aqui  # <-- CONFIGURE AQUI
DIREITO_LUX_IA_OPENAI_MODEL=gpt-4-turbo-preview

# Embeddings
DIREITO_LUX_IA_EMBEDDING_MODEL=sentence-transformers/all-MiniLM-L6-v2
DIREITO_LUX_IA_EMBEDDING_DIMENSION=384

# RAG Configuration
DIREITO_LUX_IA_RETRIEVAL_TOP_K=5
DIREITO_LUX_IA_CHUNK_SIZE=1000
DIREITO_LUX_IA_CHUNK_OVERLAP=200
DIREITO_LUX_IA_SIMILARITY_THRESHOLD=0.7

# Evaluation (Ragas)
DIREITO_LUX_IA_EVALUATION_ENABLED=true
DIREITO_LUX_IA_EVALUATION_BATCH_SIZE=10

# Observability
DIREITO_LUX_IA_METRICS_ENABLED=true
DIREITO_LUX_IA_TRACING_ENABLED=true

# Redis
DIREITO_LUX_IA_REDIS_HOST=localhost
DIREITO_LUX_IA_REDIS_PORT=6379
DIREITO_LUX_IA_REDIS_DB=2
DIREITO_LUX_IA_REDIS_PASSWORD=redis123

# Data Loss Prevention
DIREITO_LUX_IA_DLP_ENABLED=false

# Legal Knowledge Base
DIREITO_LUX_IA_KNOWLEDGE_BASE_PATH=./data/knowledge_base
DIREITO_LUX_IA_UPDATE_KNOWLEDGE_BASE=false

# Rate Limiting
DIREITO_LUX_IA_RATE_LIMIT_ENABLED=true
DIREITO_LUX_IA_RATE_LIMIT_REQUESTS=100
DIREITO_LUX_IA_RATE_LIMIT_WINDOW=3600

# Background Tasks (Celery)
DIREITO_LUX_IA_CELERY_BROKER_URL=redis://:redis123@localhost:6379/3
DIREITO_LUX_IA_CELERY_RESULT_BACKEND=redis://:redis123@localhost:6379/3

# External APIs
OPENAI_API_KEY=sk-proj-sua-chave-aqui  # <-- CONFIGURE AQUI TAMBÉM
REDIS_PASSWORD=redis123
```

## 🚀 Como Iniciar

1. **Criar o arquivo .env** (como mostrado acima)
2. **Adicionar sua chave OpenAI**
3. **Iniciar os serviços:**
   ```bash
   docker-compose -f docker-compose.ia.yml up -d
   ```

## 🧪 Teste Rápido

Para testar sem Docker:
```bash
# Instalar dependências
pip install -r requirements-simple.txt

# Executar serviço de teste
python test_service.py
```

## 📖 Documentação

- API Docs: http://localhost:9003/docs
- Postman Collection: `/postman/Direito-Lux-IA-Module.postman_collection.json`
- Guia completo: `/TESTE-RAPIDO-IA.md`

## ⚠️ Segurança

**NUNCA** commite o arquivo `.env` com chaves reais! O `.gitignore` já está configurado para ignorá-lo.