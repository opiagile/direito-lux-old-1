# 📮 Guia das Collections Postman - Direito Lux

## 📊 Collections Disponíveis

### 🔗 **1. Direito-Lux-Complete-API.postman_collection.json**
**📋 Propósito:** Collection completa para testar toda a API do sistema
- ✅ **API Principal (Go)** - Backend REST funcional
- ✅ **IA Jurídica (Python)** - Módulo de inteligência artificial
- ✅ **Ambientes:** DEV (GKE) + Local + IA

### 🤖 **2. Direito-Lux-IA-Module.postman_collection.json** 
**📋 Propósito:** Collection específica para módulo de IA jurídica
- ✅ **RAG Queries** - Consultas com Retrieval-Augmented Generation
- ✅ **Knowledge Base** - Gestão da base de conhecimento
- ✅ **Evaluation** - Avaliação de qualidade com Ragas

## 🚀 Quick Start

### **1. Importar Collections**
```bash
# No Postman, clicar em "Import"
# Selecionar os arquivos:
postman/Direito-Lux-Complete-API.postman_collection.json
postman/Direito-Lux-IA-Module.postman_collection.json
```

### **2. Configurar Environment Variables**
```javascript
// Global Variables (Settings > Variables)
base_url_dev: http://104.154.62.30
base_url_local: http://localhost:8080
base_url_ia: http://localhost:9003
jwt_token: (será preenchido automaticamente após login)
```

### **3. Testar Ambiente DEV**
```bash
# 1. Health Check
GET {{base_url_dev}}/health

# Response esperado:
{
  "status": "healthy",
  "mode": "full",
  "time": 1749687881
}
```

## 🏗️ Estrutura da Collection Completa

### **🏥 Health & Status**
- **Health Check - DEV** ✅ Funcional
- **Health Check - Local** (para desenvolvimento)

### **🔐 Authentication (em implementação)**
- **Login** - Autenticar e obter JWT
- **Refresh Token** - Renovar token
- **Forgot Password** - Recuperação de senha

### **🏢 Tenant Management (em implementação)**
- **Create Tenant** - Criar novo escritório
- **List Tenants** - Listar com paginação
- **Get Tenant** - Detalhes específicos
- **Update Tenant** - Atualizar informações
- **Tenant Usage Stats** - Estatísticas de uso

### **👤 User Profile (em implementação)**
- **Get Profile** - Perfil do usuário
- **Update Profile** - Atualizar perfil

### **🤖 IA Jurídica**
- **Health Check IA** - Status do serviço Python
- **Consulta Jurídica** - RAG queries
- **Batch Query** - Múltiplas consultas
- **Knowledge Base Stats** - Estatísticas
- **Add Legal Document** - Adicionar documentos
- **Evaluate Response** - Qualidade com Ragas

### **📊 Database & System**
- **Available Plans** - Planos disponíveis (seed data)
- **System Stats** - Estatísticas gerais

## 🌍 Ambientes e URLs

### **🧪 DEV (GKE) - ATIVO**
```
Base URL: http://104.154.62.30
Status: ✅ Funcional
Health: http://104.154.62.30/health
Features: Backend Go + PostgreSQL + Redis
```

### **💻 Local Development**
```
Backend Go: http://localhost:8080
IA Python: http://localhost:9003
Status: ⚠️ Requer setup local
Features: Desenvolvimento completo
```

### **🔬 Staging (Planejado)**
```
Base URL: https://homolog.direito-lux.com.br
Status: 📋 Não implementado
Features: Cloud SQL + Memorystore
```

### **🚀 Production (Planejado)**
```
Base URL: https://app.direito-lux.com.br
Status: 📋 Não implementado
Features: HA + SSL + Monitoring
```

## 🧪 Exemplos de Teste

### **1. Health Check Básico**
```bash
# Request
GET http://104.154.62.30/health

# Response
HTTP/1.1 200 OK
Content-Type: application/json

{
  "status": "healthy",
  "mode": "full",
  "time": 1749687881
}
```

### **2. Teste de Autenticação (quando implementado)**
```bash
# Request
POST http://104.154.62.30/api/v1/auth/login
Content-Type: application/json

{
  "email": "admin@direito-lux.com.br",
  "password": "admin123",
  "tenant_id": "uuid-tenant"
}

# Response esperado
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

### **3. Consulta IA Jurídica**
```bash
# Request
POST http://localhost:9003/api/v1/rag/query
Content-Type: application/json

{
  "question": "O que são direitos fundamentais?",
  "query_type": "legislacao",
  "evaluate_response": true
}

# Response esperado
{
  "answer": "Direitos fundamentais são...",
  "contexts": ["Art. 5º Todos são iguais..."],
  "confidence": 0.95,
  "evaluation": {
    "relevance": 0.92,
    "accuracy": 0.89
  }
}
```

## 🔧 Scripts Automáticos

### **Pre-request Scripts**
```javascript
// Auto-detect environment
const baseUrl = pm.variables.get('base_url_dev');
if (baseUrl && baseUrl.includes('104.154.62.30')) {
    console.log('Using DEV environment');
} else if (baseUrl && baseUrl.includes('localhost')) {
    console.log('Using LOCAL environment');
}
```

### **Test Scripts**
```javascript
// Auto-extract JWT token
if (pm.response.json() && pm.response.json().access_token) {
    pm.globals.set('jwt_token', pm.response.json().access_token);
    console.log('JWT token saved globally');
}

// Basic status check
pm.test('Status code is success', function () {
    pm.expect(pm.response.code).to.be.oneOf([200, 201, 202]);
});

// Response time check
pm.test('Response time is acceptable', function () {
    pm.expect(pm.response.responseTime).to.be.below(2000);
});
```

## 📝 Workflows de Teste

### **🔄 Workflow Completo (DEV)**
1. **Health Check** → Verificar se API está ativa
2. **Login** → Obter JWT token (quando implementado)
3. **List Tenants** → Testar endpoints protegidos
4. **Create Tenant** → Testar criação de dados
5. **Tenant Stats** → Testar métricas

### **🤖 Workflow IA Jurídica**
1. **Health Check IA** → Verificar serviço Python
2. **Knowledge Stats** → Ver base de conhecimento
3. **Simple Query** → Teste básico de RAG
4. **Complex Query** → Teste com filtros
5. **Evaluate Response** → Verificar qualidade

### **🧪 Workflow de Desenvolvimento**
1. **Local Health** → Verificar setup local
2. **Database Check** → Confirmar migrations
3. **API Tests** → Testar todas as rotas
4. **Performance** → Verificar tempos de resposta

## 🚨 Troubleshooting

### **❌ Connection Refused**
```bash
# Problema: Serviço não está rodando
# Solução: Verificar status dos pods
kubectl get pods
kubectl logs -f deployment/direito-lux
```

### **❌ 401 Unauthorized**
```bash
# Problema: Token inválido ou expirado
# Solução: Fazer login novamente
POST /api/v1/auth/login

# Verificar se token está sendo enviado
Authorization: Bearer {{jwt_token}}
```

### **❌ 404 Not Found**
```bash
# Problema: Endpoint não existe
# Solução: Verificar URL e versão da API
# URLs válidas:
GET /health ✅
GET /api/v1/* (quando implementado)
```

### **❌ 500 Internal Server Error**
```bash
# Problema: Erro interno da aplicação
# Solução: Verificar logs detalhados
kubectl logs deployment/direito-lux --tail=50

# Verificar banco de dados
kubectl exec -it postgres-* -- psql -U postgres -c "SELECT version();"
```

## 📈 Monitoramento de Performance

### **Métricas Importantes**
- **Response Time:** < 500ms para health checks
- **Success Rate:** > 99% para endpoints básicos
- **Error Rate:** < 1% em operação normal

### **Benchmarks**
```javascript
// Response time tests
pm.test('Health check is fast', function () {
    pm.expect(pm.response.responseTime).to.be.below(100);
});

pm.test('API calls are responsive', function () {
    pm.expect(pm.response.responseTime).to.be.below(500);
});

pm.test('IA queries are reasonable', function () {
    pm.expect(pm.response.responseTime).to.be.below(3000);
});
```

## 🔄 Atualizações das Collections

### **Como Atualizar**
1. **Pull do repositório** → `git pull origin main`
2. **Re-import no Postman** → Substituir collections existentes
3. **Update variables** → Verificar URLs atualizadas
4. **Test workflow** → Executar smoke tests

### **Versionamento**
- **v1.0** - IA Module apenas
- **v2.0** - Complete API (atual)
- **v2.1** - Com autenticação Keycloak
- **v3.0** - Com todas as features (futuro)

---

**📮 Collections Postman completas e atualizadas para todo o sistema!**

*Última atualização: 12 de Dezembro de 2024*