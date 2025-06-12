# Sistema de Migrations e Persistência de Dados

## ✅ STATUS ATUAL: BANCO FUNCIONAL E TESTADO!

**📊 Ambiente DEV:** http://104.154.62.30/health  
**🗄️ Banco:** PostgreSQL + Redis operacionais  
**🔄 Migrations:** 3 migrations executadas com sucesso  
**📈 Dados:** Planos seed criados e disponíveis  

## ✅ Resposta Rápida: OS DADOS NÃO SÃO PERDIDOS!

O sistema **NUNCA exclui o banco** durante deploys. Usamos **migrations versionadas** que apenas **adicionam/modificam** estruturas, preservando todos os dados existentes.

## 🎯 Validação em Produção

**Logs reais de execução (DEV - 12/06/2024):**
```json
{"level":"INFO","message":"Starting database migrations"}
{"level":"INFO","message":"Migration already applied","version":"001_create_initial_tables"}
{"level":"INFO","message":"Migration already applied","version":"002_add_performance_indexes"}  
{"level":"INFO","message":"Migration already applied","version":"003_seed_initial_data"}
{"level":"INFO","message":"Database migrations completed successfully"}
{"level":"INFO","message":"Database initialized successfully"}
```

## 📊 Estratégia de Persistência por Ambiente

### 🧪 **Desenvolvimento (DEV)**
- **Banco**: PostgreSQL no Kubernetes
- **Storage**: Persistent Volume (20GB)
- **Backup**: Não obrigatório (dados de teste)
- **Migrations**: Executadas automaticamente no startup

### 🔬 **Homologação (STAGING)**  
- **Banco**: Cloud SQL PostgreSQL (db-n1-standard-1, 50GB)
- **Backup**: Automático (7 dias de retenção)
- **HA**: Zona única (custo reduzido)
- **Migrations**: Versionadas com rollback

### 🚀 **Produção (PRODUCTION)**
- **Banco**: Cloud SQL PostgreSQL HA (db-n1-standard-2, 100GB+)
- **Backup**: Automático (30 dias + PITR)
- **HA**: Multi-zona com failover automático
- **Migrations**: Controladas com aprovação manual

## 🔄 Sistema de Migrations Versionadas

### Como Funciona

```go
// Cada migration tem uma versão única
type Migration struct {
    Version     string    // "001_create_initial_tables"
    Description string    // "Criar tabelas principais"
    Up          func()    // Aplicar mudanças
    Down        func()    // Reverter mudanças (rollback)
    Checksum    string    // Validação de integridade
}
```

### Fluxo de Execução

1. **Startup da aplicação** → Checa migrations pendentes
2. **Migration Manager** → Compara versões no banco vs. código
3. **Execução sequencial** → Aplica apenas migrations novas
4. **Registro no banco** → Salva versão + timestamp + checksum
5. **Log detalhado** → Auditoria completa das mudanças

### Exemplo de Migration

```go
// Migration 004: Adicionar coluna "phone" na tabela users
{
    Version: "004_add_user_phone",
    Description: "Adicionar campo telefone para usuários",
    Up: func(db *gorm.DB) error {
        return db.Exec("ALTER TABLE users ADD COLUMN phone VARCHAR(20)").Error
    },
    Down: func(db *gorm.DB) error {
        return db.Exec("ALTER TABLE users DROP COLUMN phone").Error
    },
}
```

## 🗄️ Configuração Cloud SQL

### Instância de Desenvolvimento
```yaml
db_tier: db-f1-micro        # 1 vCPU, 0.6GB RAM
disk_size: 20GB             # SSD inicial
autoresize: até 100GB       # Crescimento automático
backup: 7 dias              # Retenção básica
availability: ZONAL         # Zona única
```

### Instância de Produção
```yaml
db_tier: db-n1-standard-2   # 2 vCPU, 7.5GB RAM
disk_size: 100GB            # SSD inicial
autoresize: até 1TB         # Crescimento automático
backup: 30 dias + PITR      # Point-in-time recovery
availability: REGIONAL      # Multi-zona HA
```

## 🔐 Segurança e Acesso

### Usuários do Banco
- **direito_lux_app**: Usuário principal da aplicação (read/write)
- **direito_lux_readonly**: Usuário para relatórios e analytics
- **postgres**: Usuário administrativo (emergências)

### Conexão Segura
- **IP Privado**: Apenas dentro da VPC
- **SSL/TLS**: Obrigatório em produção
- **Cloud SQL Proxy**: Para conexões do Kubernetes
- **Workload Identity**: Autenticação sem service account keys

## 📈 Monitoramento e Alertas

### Métricas Monitoradas
- **Conexões ativas**: Previne esgotamento do pool
- **Queries lentas**: Detecção de problemas de performance
- **Utilização de disco**: Alerta antes de atingir limite
- **CPU/Memory**: Otimização de recursos

### Alertas Configurados
- **95% de utilização de disco** → Alerta crítico
- **Query > 5 segundos** → Investigação de performance
- **Falha de backup** → Alerta imediato
- **Conexões > 80%** → Scale up automático

## 🔄 Processo de Deploy

### 1. **Build & Test**
```bash
go test ./...                    # Testes unitários
docker build -t app:latest .    # Build da imagem
```

### 2. **Deploy DEV** (Automático)
```bash
kubectl apply -f postgres-dev.yaml    # PostgreSQL local
kubectl apply -f app-deployment.yaml  # Deploy da app
# → Migrations executadas automaticamente no startup
```

### 3. **Deploy STAGING** (Automático após DEV)
```bash
# Cloud SQL já existe (provisionado via Terraform)
kubectl apply -f cloud-sql-proxy.yaml # Proxy para conexão segura
helm upgrade --install direito-lux ./ # Deploy via Helm
# → Migrations executadas automaticamente
```

### 4. **Deploy PRODUCTION** (Manual com aprovação)
```bash
# Requer aprovação manual no GitHub
# Cloud SQL HA já existe
# Blue-Green deployment
helm upgrade direito-lux-green ./      # Deploy na versão green
# Health checks automáticos
# Switch de tráfego após validação
```

## 🛡️ Estratégia de Backup

### Backups Automáticos (Cloud SQL)
- **Frequência**: Diário às 3h (horário de menor uso)
- **Retenção**: 30 dias em produção, 7 dias em staging
- **PITR**: Point-in-time recovery até 7 dias atrás
- **Localização**: Multi-região para disaster recovery

### Backup Manual (Emergência)
```bash
# Export completo
gcloud sql export sql INSTANCE_NAME gs://backup-bucket/backup-$(date +%Y%m%d).sql

# Import em caso de restauração
gcloud sql import sql INSTANCE_NAME gs://backup-bucket/backup-20241106.sql
```

## ⚡ Performance e Otimização

### Conexão Pool
```go
maxOpenConns: 25        # Máximo de conexões simultâneas
maxIdleConns: 5         # Conexões idle mantidas
connMaxLifetime: 1h     # Renovação de conexões
```

### Índices Automáticos
- **Tenant ID**: Todas as consultas multi-tenant
- **Email**: Busca rápida de usuários
- **Created At**: Ordenação temporal
- **Status**: Filtros de estado

### Cache Strategy
- **Redis**: Cache de sessões e queries frequentes
- **TTL**: 5 minutos para dados dinâmicos, 1 hora para dados estáticos
- **Invalidação**: Automática quando dados são alterados

## 🎯 Comandos Úteis

### Verificar Status de Migrations
```bash
# Listar migrations aplicadas
curl http://localhost:8080/admin/migrations

# Logs da aplicação
kubectl logs -f deployment/direito-lux | grep migration
```

### Conectar ao Banco (Debug)
```bash
# Via Cloud SQL Proxy
gcloud sql connect INSTANCE_NAME --user=direito_lux_app

# Via kubectl (desenvolvimento)
kubectl exec -it postgres-pod -- psql -U postgres direito_lux
```

### Rollback de Migration (Emergência)
```bash
# Apenas em caso extremo - requer acesso direto ao banco
# Normalmente feito via nova migration que reverte as mudanças
```

## 📋 Checklist de Deploy

### ✅ Pré-Deploy
- [ ] Tests passando
- [ ] Migrations revisadas
- [ ] Backup recente confirmado
- [ ] Monitoramento ativo

### ✅ Durante Deploy
- [ ] Logs monitorados em tempo real
- [ ] Health checks validados
- [ ] Performance estável
- [ ] Erro rate normal

### ✅ Pós-Deploy
- [ ] Migrations aplicadas com sucesso
- [ ] Funcionalidades testadas
- [ ] Alertas silenciosos
- [ ] Documentação atualizada

---

## 💡 Resumo Executivo

**✅ Dados são preservados sempre**  
**✅ Migrations versionadas e auditadas**  
**✅ Backups automáticos em produção**  
**✅ Zero downtime em deploys**  
**✅ Rollback seguro quando necessário**

O sistema foi projetado para **máxima confiabilidade** e **zero perda de dados** em qualquer ambiente.