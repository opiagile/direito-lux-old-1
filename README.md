# 🏛️ Direito Lux - SaaS Jurídico

Sistema completo de gestão jurídica com IA integrada para escritórios de advocacia.

## 🚀 Módulos Implementados

### ✅ Módulo 0: Infraestrutura
- **Status:** Configurado
- **Tecnologias:** Terraform, GKE, ArgoCD, GitHub Actions
- **Ambientes:** Development, Staging, Production

### ✅ Módulo 1: Autenticação e Autorização  
- **Status:** Implementado
- **Tecnologias:** Keycloak, OPA, JWT
- **Features:** SSO, RBAC, Multi-tenant

### ✅ Módulo 2: API Gateway e Health Check
- **Status:** Implementado
- **Tecnologias:** Kong Gateway, Circuit Breaker
- **Features:** Rate limiting, Load balancing

### ✅ Módulo 3: Consulta Jurídica + Circuit Breaker + ELK
- **Status:** Implementado
- **Tecnologias:** Go, Elasticsearch, Logstash, Kibana
- **Features:** Busca jurídica, Logs centralizados

### ✅ Módulo 4: IA Jurídica (RAG + Avaliação)
- **Status:** Implementado
- **Tecnologias:** Python, FastAPI, LangChain, ChromaDB, Ragas
- **Features:** Análise de documentos, RAG, Avaliação de qualidade

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

## 🏗️ Arquitetura

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   DEV (GKE)     │    │ STAGING (GKE)   │    │  PROD (GKE)     │
│                 │    │                 │    │                 │
│ • Kong Gateway  │    │ • Kong Gateway  │    │ • Kong Gateway  │
│ • Keycloak      │    │ • Keycloak      │    │ • Keycloak      │
│ • Go Services   │    │ • Go Services   │    │ • Go Services   │
│ • Python IA     │    │ • Python IA     │    │ • Python IA     │
│ • ELK Stack     │    │ • ELK Stack     │    │ • ELK Stack     │
│ • ChromaDB      │    │ • ChromaDB      │    │ • ChromaDB      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
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

### Development
```bash
# Iniciar ambiente local
docker-compose up -d

# Acessar serviços
curl http://localhost:8080/health  # API Gateway
curl http://localhost:9002/health  # Consulta Jurídica
curl http://localhost:9003/health  # IA Jurídica
```

### Production Deploy
```bash
# Deploy via GitOps (automático)
git push origin main

# Monitor deploy
kubectl get pods -n direito-lux-prod
```

## 📝 Documentação

- [📋 Configuração de Secrets](docs/ALL-SECRETS-GUIDE.md)
- [🔧 Setup GitHub Actions](docs/GITHUB-SECRETS-SETUP.md)
- [🏗️ Infraestrutura](infrastructure/README.md)
- [🤖 IA Jurídica](services/ia-juridica/README.md)

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