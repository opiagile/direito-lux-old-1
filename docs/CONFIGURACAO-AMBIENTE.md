# 🔧 Configuração de Ambiente - Direito Lux

## 🎯 Overview

Este documento detalha como configurar e gerenciar environment variables, arquivos de configuração e deployment em diferentes ambientes.

## 📊 Precedência de Configuração

### **Ordem de Precedência (Maior → Menor)**
1. **🔥 Environment Variables** (DIREITO_LUX_*)
2. **📄 config.yaml** (se existir)
3. **⚙️ Defaults do Viper** (hardcoded)

```go
// Implementação atual em internal/config/config.go
func Load() (*Config, error) {
    // 1. Set defaults first
    setDefaults()
    
    // 2. Read config file (optional)
    viper.ReadInConfig()
    
    // 3. Environment variables override everything
    viper.AutomaticEnv()
    viper.SetEnvPrefix("DIREITO_LUX")
    
    // 4. Explicit binding for critical vars
    viper.BindEnv("database.host", "DIREITO_LUX_DATABASE_HOST")
    viper.BindEnv("redis.host", "DIREITO_LUX_REDIS_HOST")
    // ...
}
```

## 🌍 Environment Variables

### **Padrão de Nomenclatura**
- **Prefixo:** `DIREITO_LUX_`
- **Formato:** `DIREITO_LUX_SECTION_FIELD`
- **Exemplo:** `DIREITO_LUX_DATABASE_HOST=postgres`

### **Variáveis Obrigatórias**

#### **Database (PostgreSQL)**
```bash
export DIREITO_LUX_DATABASE_HOST=postgres     # Kubernetes service name
export DIREITO_LUX_DATABASE_PORT=5432
export DIREITO_LUX_DATABASE_USER=postgres
export DIREITO_LUX_DATABASE_PASSWORD=postgres123
export DIREITO_LUX_DATABASE_DBNAME=direito_lux
export DIREITO_LUX_DATABASE_SSLMODE=disable   # disable para dev
```

#### **Redis Cache**
```bash
export DIREITO_LUX_REDIS_HOST=redis           # Kubernetes service name
export DIREITO_LUX_REDIS_PORT=6379
export DIREITO_LUX_REDIS_PASSWORD=""          # Vazio para dev
export DIREITO_LUX_REDIS_DB=0
```

#### **Server Configuration**
```bash
export DIREITO_LUX_SERVER_PORT=8080
export DIREITO_LUX_SERVER_MODE=debug          # debug, release, test
export DIREITO_LUX_SERVER_READ_TIMEOUT=15s
export DIREITO_LUX_SERVER_WRITE_TIMEOUT=15s
```

### **Variáveis Opcionais**

#### **Keycloak (quando implementado)**
```bash
export DIREITO_LUX_KEYCLOAK_BASE_URL=http://keycloak:8080
export DIREITO_LUX_KEYCLOAK_REALM=direito-lux
export DIREITO_LUX_KEYCLOAK_CLIENT_ID=direito-lux-app
export DIREITO_LUX_KEYCLOAK_CLIENT_SECRET=your-secret
export DIREITO_LUX_KEYCLOAK_ADMIN_USER=admin
export DIREITO_LUX_KEYCLOAK_ADMIN_PASS=admin
```

#### **Logging**
```bash
export DIREITO_LUX_LOGGER_LEVEL=info          # debug, info, warn, error
export DIREITO_LUX_LOGGER_ENCODING=json       # json, console
export DIREITO_LUX_LOGGER_OUTPUT_PATH=stdout
```

#### **Features Toggle**
```bash
export DEMO_MODE=false                        # true para mode demo
export HEALTHCHECK_ONLY=false                # true para só health check
export ENVIRONMENT=dev                       # dev, staging, prod
```

## 📁 Arquivos de Configuração

### **config.yaml (Local Development)**
```yaml
# ATENÇÃO: Este arquivo é IGNORADO em containers (.dockerignore)
# Use apenas para desenvolvimento local

server:
  port: "8080"
  mode: "debug"

database:
  host: "localhost"        # Sobrescrito por DIREITO_LUX_DATABASE_HOST
  port: "5432"
  user: "postgres"
  password: "postgres"
  dbname: "direito_lux"
  sslmode: "disable"

redis:
  host: "localhost"        # Sobrescrito por DIREITO_LUX_REDIS_HOST
  port: "6379"
  password: ""
  db: 0

logger:
  level: "debug"           # Mais verboso para dev local
  encoding: "console"      # Mais legível para dev local
```

### **.dockerignore**
```gitignore
# Config files são ignorados no build Docker
config.yaml
config.yaml.example
config.docker.yaml

# Força uso de environment variables em containers
.env*
```

## 🚀 Configuração por Ambiente

### **🧪 Development (Local)**
```bash
# Clone e setup
git clone https://github.com/opiagile/direito-lux.git
cd direito-lux

# PostgreSQL local (Docker)
docker run -d --name postgres-dev \
  -e POSTGRES_DB=direito_lux \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 postgres:16

# Redis local (Docker)
docker run -d --name redis-dev \
  -p 6379:6379 redis:7-alpine

# Environment variables
export DIREITO_LUX_DATABASE_HOST=localhost
export DIREITO_LUX_REDIS_HOST=localhost

# Executar aplicação
go run cmd/api/main.go
```

### **☁️ Development (GKE)**
```yaml
# Atual deployment no Kubernetes
apiVersion: apps/v1
kind: Deployment
metadata:
  name: direito-lux
spec:
  template:
    spec:
      containers:
      - name: direito-lux
        env:
        - name: DIREITO_LUX_DATABASE_HOST
          value: "postgres"              # Service name
        - name: DIREITO_LUX_REDIS_HOST
          value: "redis"                 # Service name
        - name: DEMO_MODE
          value: "false"
```

### **🔬 Staging (Planejado)**
```yaml
# Cloud SQL + Memorystore
env:
- name: DIREITO_LUX_DATABASE_HOST
  value: "cloud-sql-proxy"              # Proxy para Cloud SQL
- name: DIREITO_LUX_DATABASE_SSLMODE
  value: "require"                      # SSL obrigatório
- name: DIREITO_LUX_REDIS_HOST
  value: "memorystore-redis"            # Memorystore service
- name: DIREITO_LUX_SERVER_MODE
  value: "release"                      # Modo produção
```

### **🚀 Production (Planejado)**
```yaml
# Configuração máxima segurança
env:
- name: DIREITO_LUX_DATABASE_PASSWORD
  valueFrom:
    secretKeyRef:
      name: db-credentials
      key: password
- name: DIREITO_LUX_REDIS_PASSWORD
  valueFrom:
    secretKeyRef:
      name: redis-credentials
      key: password
```

## 🔧 Debug e Troubleshooting

### **Verificar Configuração Carregada**
```go
// Logs adicionados em cmd/api/main.go para debug
logger.Info("Starting Direito Lux API",
    zap.String("db_host", cfg.Database.Host),
    zap.String("db_user", cfg.Database.User),
    zap.String("server_port", cfg.Server.Port))
```

### **Comandos de Debug**
```bash
# Ver env vars do pod
kubectl describe pod direito-lux-* | grep -A 20 "Environment:"

# Testar configuração local
DIREITO_LUX_DATABASE_HOST=test-host go run cmd/api/main.go

# Verificar se config.yaml está sendo lido
ls -la /app/config.yaml  # Não deve existir no container

# Logs da aplicação
kubectl logs -f deployment/direito-lux | grep -E "(Starting|db_host|redis)"
```

### **Problemas Comuns**

#### **1. Config.yaml sobrescrevendo env vars**
```bash
# Solução: Verificar .dockerignore
grep config.yaml .dockerignore

# Deve mostrar:
# config.yaml
# config.yaml.example
```

#### **2. Viper não lendo env vars**
```go
// Solução: Binding explícito no código
viper.BindEnv("database.host", "DIREITO_LUX_DATABASE_HOST")
viper.BindEnv("redis.host", "DIREITO_LUX_REDIS_HOST")
```

#### **3. Valores default incorretos**
```go
// Verificar setDefaults() em internal/config/config.go
viper.SetDefault("database.host", "localhost")  // Default para dev local
viper.SetDefault("redis.host", "localhost")     // Default para dev local
```

## 🧪 Testes de Configuração

### **Testes Unitários**
```go
// internal/config/config_test.go
func TestLoad(t *testing.T) {
    os.Setenv("DIREITO_LUX_DATABASE_HOST", "test-host")
    cfg, err := Load()
    assert.Equal(t, "test-host", cfg.Database.Host)
}
```

### **Testes de Integração**
```bash
# Testar precedência de env vars
DIREITO_LUX_DATABASE_HOST=env-host \
DIREITO_LUX_SERVER_PORT=9999 \
go run cmd/api/main.go

# Logs devem mostrar:
# "db_host":"env-host"
# "server_port":"9999"
```

## 📋 Checklist de Configuração

### **✅ Development Setup**
- [ ] PostgreSQL local rodando
- [ ] Redis local rodando
- [ ] Environment variables definidas
- [ ] `go run cmd/api/main.go` funcionando
- [ ] Health check respondendo

### **✅ Container Build**
- [ ] `.dockerignore` excluindo config.yaml
- [ ] Dockerfile usando multi-stage build
- [ ] Imagem final sem arquivos de config
- [ ] Env vars passadas via Kubernetes

### **✅ Kubernetes Deploy**
- [ ] ConfigMap com env vars (se necessário)
- [ ] Secrets para dados sensíveis
- [ ] Service names corretos
- [ ] Health checks configurados
- [ ] Logs estruturados funcionando

## 🔄 Processo de Update

### **1. Adicionar Nova Configuração**
```go
// 1. Adicionar campo no struct (internal/config/config.go)
type NewServiceConfig struct {
    Host string
    Port string
}

// 2. Adicionar ao Config principal
type Config struct {
    // ...
    NewService NewServiceConfig
}

// 3. Adicionar default
viper.SetDefault("newservice.host", "localhost")

// 4. Adicionar binding
viper.BindEnv("newservice.host", "DIREITO_LUX_NEWSERVICE_HOST")
```

### **2. Testar Localmente**
```bash
export DIREITO_LUX_NEWSERVICE_HOST=test-host
go run cmd/api/main.go
```

### **3. Atualizar Deployment**
```yaml
# Adicionar no deployment Kubernetes
env:
- name: DIREITO_LUX_NEWSERVICE_HOST
  value: "new-service"
```

### **4. Documentar**
```markdown
# Atualizar este documento com:
# - Nova env var na seção "Variáveis Obrigatórias"
# - Exemplo de uso
# - Notas de troubleshooting
```

---

**🔧 Configuração robusta e testada para todos os ambientes!**

*Última atualização: 12 de Junho de 2024*