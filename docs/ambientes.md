# 🚀 Ambientes do Direito Lux

O projeto está configurado para funcionar em diferentes ambientes com a estratégia ideal para cada cenário.

## 📋 Estratégia por Ambiente

### 🔧 **Desenvolvimento (Local)**
- **Infraestrutura**: Docker (Keycloak, PostgreSQL, Redis)
- **API Go**: Local (`go run`)
- **Vantagens**: Hot reload, debugging fácil, desenvolvimento rápido

### 🚀 **Produção/Homologação/Demo**
- **Tudo**: Docker (API Go + Infraestrutura)
- **Vantagens**: Ambiente idêntico à produção, isolamento completo

## 🛠️ **Como Usar**

### **Desenvolvimento**
```bash
# Iniciar ambiente de desenvolvimento
./scripts/dev-start.sh

# Em outro terminal, iniciar API Go localmente
go run cmd/demo/simple.go

# Parar ambiente
./scripts/dev-stop.sh
```

### **Produção/Demo**
```bash
# Iniciar tudo containerizado
./scripts/prod-start.sh

# Parar ambiente
./scripts/prod-stop.sh
```

## 📊 **Portas e Serviços**

### **Desenvolvimento**
| Serviço | Porto | URL | Observações |
|---------|-------|-----|-------------|
| API Go | 9001 | http://localhost:9001 | **Local** (go run) |
| Keycloak | 8080 | http://localhost:8080 | Docker |
| PostgreSQL | 5432 | localhost:5432 | Docker (exposta) |
| Redis | 6379 | localhost:6379 | Docker (exposta) |

### **Produção/Demo**
| Serviço | Porto | URL | Observações |
|---------|-------|-----|-------------|
| API Go | 9001 | http://localhost:9001 | Docker |
| Keycloak 1 | 8080 | http://localhost:8080 | Docker |
| Keycloak 2 | 8081 | http://localhost:8081 | Docker (HA) |
| Nginx | 80, 443 | http://localhost | Docker (LB) |
| PostgreSQL | - | (interno) | Docker (não exposta) |
| Redis | - | (interno) | Docker (não exposta) |

## 🔧 **Configurações**

### **Arquivos de Configuração**
- `config.yaml` - Desenvolvimento local
- `config.docker.yaml` - Ambiente Docker
- `docker-compose.dev.yml` - Infraestrutura para desenvolvimento
- `docker-compose.prod.yml` - Ambiente completo de produção

### **Credenciais Padrão**
- **Keycloak Admin**: admin/admin
- **PostgreSQL**: keycloak/keycloak
- **Banco App**: direito_lux
- **Redis**: sem senha

## 🐳 **Docker Compose Files**

### **docker-compose.dev.yml**
```yaml
# Apenas infraestrutura
services:
  - postgres (com porta exposta)
  - redis (com porta exposta) 
  - keycloak-1
```

### **docker-compose.prod.yml**
```yaml
# Ambiente completo
services:
  - postgres (interno)
  - redis (interno)
  - keycloak-1 
  - keycloak-2 (HA)
  - direito-lux-api (Docker)
  - nginx (load balancer)
```

## 🔄 **Fluxo de Desenvolvimento**

1. **Desenvolvimento diário**:
   ```bash
   ./scripts/dev-start.sh
   go run cmd/demo/simple.go
   ```

2. **Teste completo/demo**:
   ```bash
   ./scripts/prod-start.sh
   ```

3. **Deploy para produção**:
   ```bash
   # Use docker-compose.prod.yml
   # Configure variáveis de ambiente de produção
   # Configure certificados SSL
   ```

## 🚨 **Troubleshooting**

### **Problemas Comuns**

1. **Porta em uso**:
   ```bash
   lsof -ti:8080 | xargs kill -9
   ```

2. **Containers órfãos**:
   ```bash
   docker system prune -f
   ```

3. **Problemas de rede**:
   ```bash
   docker network prune -f
   ```

4. **Limpar volumes**:
   ```bash
   docker volume prune -f
   ```

### **Logs Úteis**

```bash
# Logs do ambiente de desenvolvimento
docker compose -f docker-compose.dev.yml logs -f

# Logs do ambiente de produção
docker compose -f docker-compose.prod.yml logs -f

# Logs específicos
docker logs direito-lux-keycloak-dev -f
docker logs direito-lux-api -f
```

## 📈 **Próximos Passos**

- [ ] Configurar CI/CD para build automático das imagens
- [ ] Adicionar monitoring (Prometheus/Grafana)
- [ ] Configurar certificados SSL para produção
- [ ] Adicionar backup automático do PostgreSQL
- [ ] Configurar secrets management (Vault)

## 🔗 **Links Úteis**

- **Desenvolvimento**: http://localhost:9001
- **Keycloak Admin**: http://localhost:8080/admin
- **Keycloak Account**: http://localhost:8080/realms/direito-lux/account
- **Nginx (Prod)**: http://localhost:80