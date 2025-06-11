# 🔐 Guia Completo de Secrets - Direito Lux

Este documento lista TODOS os secrets necessários para o funcionamento completo do Direito Lux.

## 📋 **Repository Secrets (Já Configurados)**

✅ **Infraestrutura GCP:**
- `GCP_SA_KEY_TERRAFORM` - Service Account para Terraform
- `GCP_SA_KEY_DEV` - Service Account para ambiente DEV
- `GCP_SA_KEY_STAGING` - Service Account para ambiente STAGING  
- `GCP_SA_KEY_PROD` - Service Account para ambiente PRODUCTION
- `GCP_PROJECT_ID_DEV` - `direito-lux-dev`
- `GCP_PROJECT_ID_STAGING` - `direito-lux-staging`
- `GCP_PROJECT_ID_PROD` - `direito-lux-prod`

## 🤖 **Secrets de IA/APIs que PRECISAM ser adicionados**

### **OpenAI API**
- **Name:** `OPENAI_API_KEY`
- **Value:** `[SUA_OPENAI_API_KEY_AQUI]` _(já configurada nos GitHub Secrets)_
- **Usado em:** Módulo 4 - IA Jurídica

### **DataJud API (Futuro - Módulo 5)**
- **Name:** `DATAJUD_API_KEY`
- **Value:** `[Será fornecida quando implementarmos]`
- **Name:** `DATAJUD_USERNAME`
- **Value:** `[Será fornecido quando implementarmos]`
- **Name:** `DATAJUD_PASSWORD`
- **Value:** `[Será fornecida quando implementarmos]`

### **ChromaDB Credentials**
- **Name:** `CHROMA_AUTH_TOKEN`
- **Value:** `direito-lux-chroma-token-2024` _(gerado automaticamente)_

## 🔐 **Secrets de Banco de Dados**

### **PostgreSQL DEV**
- **Name:** `DB_DEV_HOST`
- **Value:** `127.0.0.1`
- **Name:** `DB_DEV_PORT`
- **Value:** `5432`
- **Name:** `DB_DEV_NAME`
- **Value:** `direito_lux_dev`
- **Name:** `DB_DEV_USERNAME`
- **Value:** `postgres`
- **Name:** `DB_DEV_PASSWORD`
- **Value:** `DireitoLux@DevDB2024!`

### **Redis DEV**
- **Name:** `REDIS_DEV_HOST`
- **Value:** `127.0.0.1`
- **Name:** `REDIS_DEV_PORT`
- **Value:** `6379`
- **Name:** `REDIS_DEV_PASSWORD`
- **Value:** `DireitoLux@Redis2024!`

## 🔑 **Secrets de Autenticação**

### **Keycloak Admin**
- **Name:** `KEYCLOAK_ADMIN_USERNAME`
- **Value:** `admin`
- **Name:** `KEYCLOAK_ADMIN_PASSWORD`
- **Value:** `DireitoLux@2024!Admin`

### **JWT Secrets**
- **Name:** `JWT_SECRET_KEY`
- **Value:** `4f8a9d2e7c1b6f5a3e8d9c2b7f4a1e6d9c3b8f2a5e7d1c4b9f6a3e8d2c7b5f1a4e9d`
- **Name:** `JWT_EXPIRATION`
- **Value:** `24h`

## 📞 **Secrets de Notificação**

### **Slack Webhook**
- **Name:** `SLACK_WEBHOOK`
- **Value:** `https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK`

### **Email SMTP (Futuro)**
- **Name:** `SMTP_HOST`
- **Value:** `smtp.gmail.com`
- **Name:** `SMTP_PORT`
- **Value:** `587`
- **Name:** `SMTP_USERNAME`
- **Value:** `noreply@direito-lux.com.br`
- **Name:** `SMTP_PASSWORD`
- **Value:** `[Será configurada]`

## 🌍 **Environment-Specific Secrets**

### **Environment: development**
- **Name:** `APP_ENV`
- **Value:** `development`
- **Name:** `DEBUG_MODE`
- **Value:** `true`
- **Name:** `LOG_LEVEL`
- **Value:** `debug`

### **Environment: staging**
- **Name:** `APP_ENV`
- **Value:** `staging`
- **Name:** `DEBUG_MODE`
- **Value:** `true`
- **Name:** `LOG_LEVEL`
- **Value:** `info`

### **Environment: production**
- **Name:** `APP_ENV`
- **Value:** `production`
- **Name:** `DEBUG_MODE`
- **Value:** `false`
- **Name:** `LOG_LEVEL`
- **Value:** `warn`
- **Name:** `PROD_DB_CONNECTION`
- **Value:** `postgres://user:pass@host:5432/db?sslmode=require`

## 🔗 **Secrets de Integração Externa**

### **Infracost (Monitoramento de Custos)**
- **Name:** `INFRACOST_API_KEY`
- **Value:** `[Registre em https://www.infracost.io/]`

### **GitHub Token (para ArgoCD)**
- **Name:** `GITHUB_TOKEN`
- **Value:** `[Seu Personal Access Token]`

## 🚨 **Secrets de Segurança**

### **Encryption Keys**
- **Name:** `DATA_ENCRYPTION_KEY`
- **Value:** `8b5f2a7e1d4c9b6f3a8e5d2c7b4f1a9e6d3c8b5f2a7e1d4c9b6f3a8e5d2c7b4f1a`
- **Name:** `SESSION_SECRET`
- **Value:** `direito-lux-session-secret-key-ultra-secure-2024-production-ready`

### **Rate Limiting**
- **Name:** `RATE_LIMIT_SECRET`
- **Value:** `direito-lux-rate-limit-secret-2024`

## 📝 **Como Adicionar no GitHub**

1. **Repository Secrets:** GitHub → Settings → Secrets and variables → Actions
2. **Environment Secrets:** GitHub → Settings → Environments → [dev/staging/prod] → Add secret

## 🎯 **Prioridade de Implementação**

### **🔴 CRÍTICO (Adicionar AGORA):**
- ✅ `OPENAI_API_KEY` - Já configurada
- ✅ `JWT_SECRET_KEY` - Já configurada
- ✅ `DATA_ENCRYPTION_KEY` - Já configurada

### **🟡 IMPORTANTE (Próximas semanas):**
- `SLACK_WEBHOOK` - Para notificações CI/CD
- `KEYCLOAK_ADMIN_PASSWORD` - Para segurança
- Database passwords - Para produção

### **🟢 FUTURO (Quando implementar):**
- `DATAJUD_API_KEY` - Módulo 5
- `SMTP_*` - Sistema de emails
- `GITHUB_TOKEN` - ArgoCD automático

## 🛡️ **Boas Práticas de Segurança**

1. **Nunca commitar secrets no código**
2. **Rotacionar chaves mensalmente**
3. **Usar diferentes chaves por ambiente**
4. **Logs não devem mostrar secrets**
5. **Monitoring de uso de APIs**

## 🔄 **Script de Rotação de Chaves**

```bash
# Futuro: Script automático para rotacionar chaves
./scripts/rotate-secrets.sh --environment=prod --type=all
```