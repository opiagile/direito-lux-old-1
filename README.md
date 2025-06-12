# 🏛️ Direito Lux - SaaS Jurídico Enterprise

Sistema completo de gestão jurídica com IA integrada, migrations versionadas e arquitetura de microsserviços escalável.

## 📊 Status Atual do Projeto

**🎯 Ambiente DEV:** ✅ **FUNCIONAL E ACESSÍVEL**
- **URL:** http://104.154.62.30/health
- **Banco:** PostgreSQL + Redis funcionando
- **API:** Todas as rotas ativas
- **Migrations:** 3 migrations executadas com sucesso
- **Pipeline:** CI/CD funcionando perfeitamente

## 🚀 Módulos Implementados

### ✅ Módulo 0: Infraestrutura & CI/CD
- **Status:** ✅ **COMPLETO**
- **Tecnologias:** Terraform, GKE, ArgoCD, GitHub Actions
- **Ambientes:** Development, Staging, Production
- **Features:** GitOps, IaC, Multi-ambiente, Cost monitoring

### ✅ Módulo 1: Core Backend + Auth
- **Status:** ✅ **COMPLETO**
- **Tecnologias:** Go 1.21, Gin, GORM, Keycloak, JWT
- **Features:** REST API, Multi-tenant, RBAC, Circuit Breaker
- **Banco:** PostgreSQL com migrations versionadas

### ✅ Módulo 2: API Gateway + Monitoring  
- **Status:** ✅ **COMPLETO**
- **Tecnologias:** Kong Gateway, OPA, Prometheus, Grafana
- **Features:** Rate limiting, Load balancing, Health checks

### ✅ Módulo 3: Sistema de Consultas + Observabilidade
- **Status:** ✅ **COMPLETO** 
- **Tecnologias:** Go, ELK Stack, OpenTelemetry, Jaeger
- **Features:** Logs centralizados, Tracing distribuído, Alertas

### ✅ Módulo 4: IA Jurídica (RAG + Avaliação)
- **Status:** ✅ **COMPLETO**
- **Tecnologias:** Python 3.11, FastAPI, LangChain, ChromaDB, Ragas
- **Features:** RAG jurídico, Avaliação contínua, Templates de prompts
- **APIs:** Consulta IA, Batch processing, Knowledge management

### 🔄 Módulo 5: Mensageria + DataJud (Em desenvolvimento)
- **Status:** Planejado
- **Tecnologias:** WhatsApp Business API, Telegram Bot, DataJud API
- **Features:** Consulta via chat, Integração judicial

### 📅 Módulo 6: Admin Panel Web (Futuro)
- **Status:** Planejado
- **Tecnologias:** React, TypeScript, Material-UI
- **Features:** Dashboard, Relatórios, Configurações

### 💰 Módulo 7: Billing e Reports (Futuro)
- **Status:** Planejado
- **Tecnologias:** Stripe, PDF Reports, Analytics
- **Features:** Cobrança, Relatórios financeiros

## 🏗️ Arquitetura Atual (Dezembro 2024)

### 🌐 **Ambiente DEV (GKE Funcional)**
```
┌─────────────────────────────────────────────────────────────┐
│                    GKE Cluster (us-central1-a)              │
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │   Direito-Lux   │  │   PostgreSQL    │  │    Redis     │ │
│  │   (Go API)      │  │   (Database)    │  │   (Cache)    │ │
│  │                 │  │                 │  │              │ │
│  │ • REST API      │  │ • 3 Migrations  │  │ • Sessions   │ │
│  │ • Multi-tenant  │  │ • Seed Data     │  │ • Cache      │ │
│  │ • Health Checks │  │ • Audit Logs    │  │              │ │
│  │ • Circuit Break │  │                 │  │              │ │
│  └─────────────────┘  └─────────────────┘  └──────────────┘ │
│                                                             │
│  External IP: 104.154.62.30                                │
│  Health: http://104.154.62.30/health                       │
└─────────────────────────────────────────────────────────────┘
```

### 🔄 **CI/CD Pipeline Ativo**
```
GitHub Push → GitHub Actions → Docker Build → GKE Deploy
     ↓              ↓               ↓             ↓
   Commit      [Build/Test]    [Artifact Reg]  [Rolling Update]
              [Security Scan]   [Multi-Stage]   [Health Check]
              [Go Test/Lint]    [Optimized]     [Zero Downtime]
```

### 💾 **Banco de Dados (PostgreSQL)**
```sql
-- Tabelas Criadas:
✅ migration_versions  -- Controle de migrations
✅ tenants            -- Multi-tenancy
✅ plans              -- Planos de assinatura (3 criados)
✅ subscriptions      -- Assinaturas ativas
✅ users              -- Usuários do sistema
✅ audit_logs         -- Logs de auditoria
✅ api_keys           -- Chaves de API

-- Dados Seed:
✅ 3 Planos: starter (R$99), professional (R$299), enterprise (R$999)
✅ Índices de performance criados
✅ Constraints e relações configuradas
```

## 🔄 CI/CD Pipeline

- **GitHub Actions:** Build, Test, Security Scan
- **ArgoCD:** GitOps deployment
- **Terraform:** Infrastructure as Code
- **Environments:** dev → staging → production

## 🔐 Segurança

- **Secrets Management:** GitHub Secrets + Environment protection
- **API Security:** JWT, OAuth2, Rate limiting
- **Data Protection:** Encryption at rest, DLP policies
- **Network Security:** VPC, Private clusters, Network policies

## 📊 Monitoramento

- **Logs:** ELK Stack (Elasticsearch, Logstash, Kibana)
- **Metrics:** Prometheus + Grafana
- **Alerting:** Slack notifications
- **Cost Monitoring:** Infracost integration

## 🚀 Quick Start

### 🌐 **Acesso Imediato (Ambiente DEV)**
```bash
# Health Check da API
curl http://104.154.62.30/health

# Response esperado:
{
  "status": "healthy",
  "mode": "full", 
  "time": 1749687881
}
```

### 🛠️ **Desenvolvimento Local**
```bash
# Clone do repositório
git clone https://github.com/opiagile/direito-lux.git
cd direito-lux

# Configurar environment variables
export DIREITO_LUX_DATABASE_HOST=localhost
export DIREITO_LUX_DATABASE_USER=postgres
export DIREITO_LUX_DATABASE_PASSWORD=postgres

# Build e execução
go mod tidy
go run cmd/api/main.go

# Ou via Docker
docker build -t direito-lux .
docker run -p 8080:8080 direito-lux
```

### 🚀 **Deploy Automático**
```bash
# Deploy via GitOps (automático)
git add .
git commit -m "feat: nova feature"
git push origin main

# Monitorar deploy
kubectl get pods --watch
kubectl logs -f deployment/direito-lux

# Verificar saúde
curl http://104.154.62.30/health
```

### 🧪 **Testes e Qualidade**
```bash
# Executar todos os testes
go test ./...

# Verificar formatação
gofmt -l .

# Verificar problemas
go vet ./...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📝 Documentação Completa

### **🚀 Getting Started**
- [📊 Setup Atual - Status do Projeto](docs/SETUP-ATUAL.md)
- [🔧 Configuração de Ambiente](docs/CONFIGURACAO-AMBIENTE.md)
- [📚 Histórico de Implementação](docs/HISTORICO-IMPLEMENTACAO.md)

### **🔌 API & Development**
- [🔌 API Reference - Endpoints REST](docs/API-REFERENCE.md)
- [🗄️ Migrations e Persistência](docs/MIGRATIONS-E-PERSISTENCIA.md)
- [🧪 Testes Automatizados](internal/config/config_test.go)

### **🏗️ Infrastructure & DevOps**
- [📋 Configuração de Secrets GCP](docs/ALL-SECRETS-GUIDE.md)
- [🔧 Setup GitHub Actions](docs/GITHUB-SECRETS-SETUP.md)
- [☁️ Infraestrutura Terraform](infrastructure/terraform/)
- [🐳 Docker & Kubernetes](k8s/)

### **🤖 Módulos Específicos**
- [🤖 IA Jurídica - RAG & LangChain](services/ia-juridica/README.md)
- [⚖️ Módulo Legal - Consultas](internal/services/)
- [🔐 Autenticação - Keycloak](internal/auth/)

### **📊 Monitoring & Operations**
- [📈 Observabilidade - Logs & Metrics](infrastructure/prometheus/)
- [🚨 Alertas e Monitoramento](infrastructure/grafana/)
- [🔍 Troubleshooting Guide](docs/SETUP-ATUAL.md#-troubleshooting-guide)

## 🎯 Próximos Passos

1. **Módulo 5:** DataJud + Mensageria (WhatsApp/Telegram)
2. **Módulo 6:** Admin Panel Web (React)
3. **Módulo 7:** Billing e Reports
4. **Escalabilidade:** Multi-tenant, Auto-scaling
5. **Compliance:** LGPD, ISO 27001

## 📞 Suporte

- **Issues:** GitHub Issues
- **Docs:** `/docs` folder
- **Slack:** #direito-lux-dev

---

**🎉 Direito Lux - Transformando o futuro jurídico com IA!**

---

## 🧪 CI/CD Test

Este commit testa o pipeline CI/CD completo com GitHub Actions + ArgoCD.

**Status:** ✅ Secrets configurados, Environments criados, Pipeline ativo