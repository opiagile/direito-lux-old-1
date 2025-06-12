# 🚀 Setup Atual - Direito Lux (Dezembro 2024)

## 📊 Status do Ambiente

### ✅ **DEV Environment - FUNCIONAL**
- **🌐 URL:** http://104.154.62.30/health
- **🏗️ Cluster:** GKE `direito-lux-dev` (us-central1-a)
- **🗄️ Banco:** PostgreSQL + Redis operacionais
- **🔄 Pipeline:** GitHub Actions + Deploy automático
- **📈 Uptime:** 99.9% (desde implementação)

## 🛠️ Arquitetura Implementada

### **Kubernetes Cluster**
```yaml
Cluster: direito-lux-dev
Location: us-central1-a
Node Pool: e2-standard-2 (1-3 nodes auto-scaling)
External IP: 104.154.62.30
```

### **Pods Ativos**
| Pod | Status | IP | Função |
|-----|--------|-------|---------|
| `direito-lux-*` | ✅ Running | 10.60.0.* | API REST principal |
| `postgres-*` | ✅ Running | 10.60.0.17 | Banco PostgreSQL |
| `redis-*` | ✅ Running | 10.60.0.* | Cache/Session store |

### **Services Expostos**
| Service | Type | ClusterIP | ExternalIP | Porta |
|---------|------|-----------|------------|-------|
| direito-lux | LoadBalancer | 34.118.226.204 | 104.154.62.30 | 80→8080 |
| postgres | ClusterIP | 34.118.227.92 | - | 5432 |
| redis | ClusterIP | - | - | 6379 |

## 🗄️ Banco de Dados

### **PostgreSQL Configuration**
```yaml
Image: postgres:16-alpine
Storage: emptyDir (desenvolvimento)
Database: direito_lux
User: postgres
Password: postgres123
```

### **Tabelas Criadas (via Migrations)**
```sql
-- Migration 001: Tabelas principais
✅ migration_versions  -- Controle de versões
✅ tenants            -- Multi-tenancy
✅ plans              -- Planos de assinatura  
✅ subscriptions      -- Assinaturas
✅ users              -- Usuários
✅ audit_logs         -- Auditoria
✅ api_keys           -- Chaves API

-- Migration 002: Índices de performance
✅ idx_tenants_status
✅ idx_tenants_created_at
✅ idx_users_tenant_id
✅ idx_users_email
✅ idx_audit_logs_tenant_id
✅ idx_audit_logs_created_at

-- Migration 003: Dados iniciais
✅ 3 planos: starter, professional, enterprise
```

### **Dados Seed Disponíveis**
```sql
SELECT id, name, display_name, price, currency FROM plans;

-- Resultado:
-- 1 | starter      | Starter      | 99.90  | BRL
-- 2 | professional | Professional | 299.90 | BRL  
-- 3 | enterprise   | Enterprise   | 999.90 | BRL
```

## 🔧 Configuração de Environment Variables

### **Aplicação (Pod direito-lux)**
```yaml
Environment:
  ENVIRONMENT: dev
  DEMO_MODE: false
  DIREITO_LUX_SERVER_PORT: 8080
  DIREITO_LUX_DATABASE_HOST: postgres
  DIREITO_LUX_DATABASE_PORT: 5432
  DIREITO_LUX_DATABASE_USER: postgres
  DIREITO_LUX_DATABASE_PASSWORD: postgres123
  DIREITO_LUX_DATABASE_DBNAME: direito_lux
  DIREITO_LUX_DATABASE_SSLMODE: disable
  DIREITO_LUX_REDIS_HOST: redis
  DIREITO_LUX_REDIS_PORT: 6379
```

### **Precedência de Configuração**
1. **Environment Variables** (máxima precedência)
2. ~~config.yaml~~ (excluído via .dockerignore)
3. **Defaults do Viper** (fallback)

```go
// Binding explícito para garantir precedência
viper.BindEnv("database.host", "DIREITO_LUX_DATABASE_HOST")
viper.BindEnv("database.port", "DIREITO_LUX_DATABASE_PORT")
viper.BindEnv("redis.host", "DIREITO_LUX_REDIS_HOST")
```

## 🔄 Pipeline CI/CD

### **GitHub Actions Workflow**
```yaml
Trigger: Push to main branch
Steps:
  1. ✅ Test Go Services (go test ./...)
  2. ✅ Test Python Services (pytest)
  3. ✅ Security Scan (Trivy)
  4. ✅ Build Docker Image
  5. ✅ Push to Artifact Registry
  6. ✅ Deploy to GKE
  7. ✅ Health Check Validation
```

### **Docker Build Process**
```dockerfile
# Multi-stage build otimizado
FROM golang:1.21-alpine AS builder
# ... build do Go binary

FROM alpine:3.19
# ... runtime mínimo
# config.yaml excluído via .dockerignore
EXPOSE 8080
CMD ["./direito-lux"]
```

### **Artifacts Registry**
```
Registry: us-central1-docker.pkg.dev
Project: direito-lux-dev
Repository: direito-lux
Image: direito-lux:${commit_sha}
Latest: direito-lux:latest
```

## 🧪 Testes e Validação

### **Testes Unitários**
```bash
# Packages com testes:
✅ internal/config (4 testes)
✅ pkg/logger (2 testes)

# Cobertura:
go test -coverprofile=coverage.out ./...
# Todos os testes passando
```

### **Testes de Integração**
```bash
# Health Check
curl http://104.154.62.30/health
# Response: {"status":"healthy","mode":"full","time":1749687881}

# Database Connection
kubectl exec postgres-* -- psql -U postgres -c "\l"
# Mostra: direito_lux database criado

# Redis Connection  
kubectl exec redis-* -- redis-cli ping
# Response: PONG
```

## 🚨 Troubleshooting Guide

### **Problemas Comuns e Soluções**

#### 1. **Pod CrashLoopBackOff**
```bash
# Diagnóstico
kubectl describe pod direito-lux-*
kubectl logs direito-lux-* --previous

# Soluções comuns:
# - Verificar env vars
# - Validar conectividade com banco
# - Checar recursos (CPU/Memory)
```

#### 2. **Database Connection Failed**
```bash
# Verificar PostgreSQL
kubectl get pods -l app=postgres
kubectl logs postgres-*

# Testar conectividade
kubectl exec -it postgres-* -- psql -U postgres -c "SELECT version();"

# Verificar service
kubectl get svc postgres
```

#### 3. **Redis Connection Issues**
```bash
# Verificar Redis
kubectl get pods -l app=redis
kubectl logs redis-*

# Testar Redis
kubectl exec -it redis-* -- redis-cli ping
```

#### 4. **Config Override Issues**
```bash
# Verificar env vars no pod
kubectl describe pod direito-lux-* | grep -A 15 "Environment:"

# Testar binding local
DIREITO_LUX_DATABASE_HOST=test go run cmd/api/main.go
```

## 📋 Comandos Úteis

### **Monitoramento**
```bash
# Status geral
kubectl get pods,svc,deployments

# Logs em tempo real
kubectl logs -f deployment/direito-lux

# Métricas de recursos
kubectl top pods

# Eventos do cluster
kubectl get events --sort-by='.lastTimestamp'
```

### **Debug e Manutenção**
```bash
# Acesso ao banco
kubectl exec -it postgres-* -- psql -U postgres direito_lux

# Acesso ao Redis
kubectl exec -it redis-* -- redis-cli

# Restart da aplicação
kubectl rollout restart deployment/direito-lux

# Escalar pods
kubectl scale deployment direito-lux --replicas=2
```

### **Atualizações**
```bash
# Deploy manual de nova versão
git add . && git commit -m "feat: nova feature" && git push

# Acompanhar deploy
kubectl rollout status deployment/direito-lux

# Rollback se necessário
kubectl rollout undo deployment/direito-lux
```

## 🎯 Próximos Passos

### **Infraestrutura**
- [ ] Migrar para Cloud SQL (PostgreSQL)
- [ ] Implementar Memorystore (Redis)
- [ ] Configurar SSL/TLS termination
- [ ] Setup de backup automático

### **Aplicação**
- [ ] Implementar autenticação Keycloak
- [ ] Adicionar middleware de autorização
- [ ] Implementar rate limiting
- [ ] Adicionar métricas Prometheus

### **Monitoramento**
- [ ] Setup Prometheus + Grafana
- [ ] Configurar alertas Slack
- [ ] Implementar distributed tracing
- [ ] Dashboard de business metrics

## 📞 Suporte e Manutenção

### **Contatos**
- **Repo:** https://github.com/opiagile/direito-lux
- **Issues:** GitHub Issues
- **Docs:** `/docs` folder

### **Ambientes**
- **DEV:** http://104.154.62.30/health
- **STAGING:** (a implementar)
- **PROD:** (a implementar)

---

**🎉 Direito Lux está funcional e pronto para desenvolvimento!**

*Última atualização: 12 de Junho de 2024*