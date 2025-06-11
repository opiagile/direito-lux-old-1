# 🔐 GitHub Secrets - Guia de Configuração

Este guia te ajuda a configurar todos os secrets necessários para o CI/CD do Direito Lux funcionar.

## 📋 **Pré-requisitos**

1. ✅ Conta Google Cloud Platform ativa
2. ✅ Projetos GCP criados (direito-lux-dev, direito-lux-staging, direito-lux-prod)
3. ✅ Repositório GitHub com permissões de admin
4. ✅ gcloud CLI instalado e autenticado

## 🚀 **Passo 1: Criar Service Accounts no GCP**

### **1.1 Service Account para Terraform**

```bash
# Criar service account para Terraform
gcloud iam service-accounts create terraform-sa \
    --description="Service Account para Terraform" \
    --display-name="Terraform SA"

# Dar permissões necessárias
gcloud projects add-iam-policy-binding direito-lux-dev \
    --member="serviceAccount:terraform-sa@direito-lux-dev.iam.gserviceaccount.com" \
    --role="roles/editor"

gcloud projects add-iam-policy-binding direito-lux-dev \
    --member="serviceAccount:terraform-sa@direito-lux-dev.iam.gserviceaccount.com" \
    --role="roles/iam.serviceAccountAdmin"

# Criar chave JSON
gcloud iam service-accounts keys create terraform-key.json \
    --iam-account=terraform-sa@direito-lux-dev.iam.gserviceaccount.com
```

### **1.2 Service Account para DEV**

```bash
# Criar service account para DEV
gcloud iam service-accounts create github-actions-dev \
    --project=direito-lux-dev \
    --description="GitHub Actions DEV" \
    --display-name="GitHub Actions DEV"

# Permissões para GKE e deploy
gcloud projects add-iam-policy-binding direito-lux-dev \
    --member="serviceAccount:github-actions-dev@direito-lux-dev.iam.gserviceaccount.com" \
    --role="roles/container.developer"

gcloud projects add-iam-policy-binding direito-lux-dev \
    --member="serviceAccount:github-actions-dev@direito-lux-dev.iam.gserviceaccount.com" \
    --role="roles/storage.admin"

# Criar chave
gcloud iam service-accounts keys create github-dev-key.json \
    --iam-account=github-actions-dev@direito-lux-dev.iam.gserviceaccount.com \
    --project=direito-lux-dev
```

### **1.3 Service Account para STAGING**

```bash
# Criar service account para STAGING
gcloud iam service-accounts create github-actions-staging \
    --project=direito-lux-staging \
    --description="GitHub Actions STAGING" \
    --display-name="GitHub Actions STAGING"

# Permissões
gcloud projects add-iam-policy-binding direito-lux-staging \
    --member="serviceAccount:github-actions-staging@direito-lux-staging.iam.gserviceaccount.com" \
    --role="roles/container.developer"

gcloud projects add-iam-policy-binding direito-lux-staging \
    --member="serviceAccount:github-actions-staging@direito-lux-staging.iam.gserviceaccount.com" \
    --role="roles/storage.admin"

# Criar chave
gcloud iam service-accounts keys create github-staging-key.json \
    --iam-account=github-actions-staging@direito-lux-staging.iam.gserviceaccount.com \
    --project=direito-lux-staging
```

### **1.4 Service Account para PRODUCTION**

```bash
# Criar service account para PRODUCTION
gcloud iam service-accounts create github-actions-prod \
    --project=direito-lux-prod \
    --description="GitHub Actions PROD" \
    --display-name="GitHub Actions PROD"

# Permissões
gcloud projects add-iam-policy-binding direito-lux-prod \
    --member="serviceAccount:github-actions-prod@direito-lux-prod.iam.gserviceaccount.com" \
    --role="roles/container.developer"

gcloud projects add-iam-policy-binding direito-lux-prod \
    --member="serviceAccount:github-actions-prod@direito-lux-prod.iam.gserviceaccount.com" \
    --role="roles/storage.admin"

# Criar chave
gcloud iam service-accounts keys create github-prod-key.json \
    --iam-account=github-actions-prod@direito-lux-prod.iam.gserviceaccount.com \
    --project=direito-lux-prod
```

## 🔑 **Passo 2: Configurar Repository Secrets**

Vá para: **GitHub → Seu Repositório → Settings → Secrets and variables → Actions**

### **2.1 GCP Service Account Keys**

```bash
# Copiar conteúdo dos arquivos JSON (remover quebras de linha)
cat terraform-key.json | tr -d '\n'
cat github-dev-key.json | tr -d '\n'
cat github-staging-key.json | tr -d '\n'
cat github-prod-key.json | tr -d '\n'
```

**Adicionar no GitHub:**

| Secret Name | Value |
|-------------|--------|
| `GCP_SA_KEY_TERRAFORM` | Conteúdo de `terraform-key.json` |
| `GCP_SA_KEY_DEV` | Conteúdo de `github-dev-key.json` |
| `GCP_SA_KEY_STAGING` | Conteúdo de `github-staging-key.json` |
| `GCP_SA_KEY_PROD` | Conteúdo de `github-prod-key.json` |

### **2.2 Project IDs**

| Secret Name | Value |
|-------------|--------|
| `GCP_PROJECT_ID_DEV` | `direito-lux-dev` |
| `GCP_PROJECT_ID_STAGING` | `direito-lux-staging` |
| `GCP_PROJECT_ID_PROD` | `direito-lux-prod` |

### **2.3 Slack Webhook (Opcional)**

1. **Criar Slack App:**
   - Vá para https://api.slack.com/apps
   - Criar nova app → From scratch
   - Nome: "Direito Lux CI/CD"
   - Workspace: Seu workspace

2. **Configurar Incoming Webhooks:**
   - Incoming Webhooks → Activate
   - Add New Webhook to Workspace
   - Escolher canal (ex: #deployments)
   - Copiar Webhook URL

| Secret Name | Value |
|-------------|--------|
| `SLACK_WEBHOOK` | `https://hooks.slack.com/services/...` |

### **2.4 Infracost API Key (Opcional)**

1. **Registrar no Infracost:**
   - Vá para https://www.infracost.io/
   - Sign up gratuito
   - Copiar API key do dashboard

| Secret Name | Value |
|-------------|--------|
| `INFRACOST_API_KEY` | Sua API key do Infracost |

## 🌍 **Passo 3: Configurar Environment Secrets**

### **3.1 Criar Environments**

Vá para: **GitHub → Settings → Environments**

Criar 3 environments:
- `development` (sem proteção)
- `staging` (com reviewers opcionais)
- `production` (com reviewers obrigatórios)

### **3.2 Environment: development**

| Secret Name | Value |
|-------------|--------|
| _Nenhum adicional necessário_ | |

### **3.3 Environment: staging**

| Secret Name | Value |
|-------------|--------|
| _Nenhum adicional necessário_ | |

### **3.4 Environment: production**

| Secret Name | Value |
|-------------|--------|
| `PROD_DB_CONNECTION` | `postgres://user:pass@host:5432/db?sslmode=require` |

**⚠️ Configurar Protection Rules:**
- ✅ Required reviewers: Adicionar você e outros devs sênior
- ✅ Wait timer: 5 minutos
- ✅ Deployment branches: Only protected branches

## 📱 **Passo 4: Configurar Notificações Slack**

### **4.1 Criar Canal #deployments**

```
/invite @direito-lux-ci-cd
```

### **4.2 Configurar Mensagens**

O webhook já está configurado nos workflows para enviar:
- ✅ Deploy DEV completo
- ⚠️ Deploy STAGING com testes
- 🚀 Deploy PRODUCTION
- ❌ Falhas em qualquer ambiente

## 🧪 **Passo 5: Testar a Configuração**

### **5.1 Script de Teste**

```bash
#!/bin/bash
# scripts/test-github-secrets.sh

echo "🧪 Testando configuração GitHub Secrets..."

# Verificar se os arquivos de chave existem
for key in terraform-key.json github-dev-key.json github-staging-key.json github-prod-key.json; do
    if [[ -f "$key" ]]; then
        echo "✅ $key existe"
    else
        echo "❌ $key não encontrado"
    fi
done

# Testar autenticação GCP
echo "🔍 Testando autenticação GCP..."

for project in direito-lux-dev direito-lux-staging direito-lux-prod; do
    if gcloud projects describe $project --quiet >/dev/null 2>&1; then
        echo "✅ Projeto $project acessível"
    else
        echo "❌ Projeto $project inacessível"
    fi
done

echo "🎯 Teste manual: Faça um commit pequeno e verifique se o pipeline executa"
```

### **5.2 Primeiro Teste**

```bash
# Fazer um commit pequeno para testar
echo "# Test" >> README.md
git add README.md
git commit -m "test: Testar pipeline CI/CD"
git push origin main
```

**Verificar:**
1. ✅ GitHub Actions executou
2. ✅ Build passou
3. ✅ Deploy DEV funcionou
4. ✅ Notificação Slack recebida

## 🔒 **Passo 6: Segurança e Limpeza**

### **6.1 Limpar Chaves Locais**

```bash
# IMPORTANTE: Remover chaves do seu computador
rm -f terraform-key.json
rm -f github-dev-key.json  
rm -f github-staging-key.json
rm -f github-prod-key.json

# Verificar que foram removidas
ls -la *.json
```

### **6.2 Rotacionar Chaves (Recomendado mensalmente)**

```bash
# Script para rotacionar chaves
scripts/rotate-service-account-keys.sh
```

## ✅ **Checklist Final**

- [ ] Service Accounts criados nos 3 projetos GCP
- [ ] Chaves JSON baixadas e adicionadas aos GitHub Secrets
- [ ] Project IDs configurados
- [ ] Slack webhook configurado (opcional)
- [ ] Environments GitHub configurados com proteções
- [ ] Teste manual executado com sucesso
- [ ] Chaves locais removidas do computador
- [ ] Time notificado sobre novos deployments automáticos

## 🚨 **Troubleshooting**

### **Erro: "Permission denied"**
```bash
# Verificar permissões do service account
gcloud projects get-iam-policy direito-lux-dev \
    --flatten="bindings[].members" \
    --filter="bindings.members:github-actions-dev*"
```

### **Erro: "Invalid key format"**
```bash
# Verificar formato JSON
cat github-dev-key.json | jq .
```

### **Erro: "Environment not found"**
- Verificar se environments foram criados no GitHub
- Verificar se nomes estão exatos (case-sensitive)

### **Slack não recebe notificações**
- Verificar se webhook URL está correto
- Testar webhook manualmente:
```bash
curl -X POST -H 'Content-type: application/json' \
    --data '{"text":"Teste do Direito Lux CI/CD"}' \
    YOUR_SLACK_WEBHOOK_URL
```

## 📞 **Suporte**

Se tiver problemas:
1. Verificar logs do GitHub Actions
2. Conferir este checklist
3. Testar service accounts manualmente
4. Rotacionar chaves se necessário

🎉 **Parabéns! Sua infraestrutura está pronta para deploy automático!**