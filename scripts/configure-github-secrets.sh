#!/bin/bash

# Script para configurar GitHub Secrets - Direito Lux
# Execute após instalar gcloud CLI

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Funções de log
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Verificar se gcloud está instalado
check_gcloud() {
    log_info "🔍 Verificando instalação do gcloud..."
    
    if ! command -v gcloud &> /dev/null; then
        log_error "gcloud CLI não encontrado!"
        echo ""
        echo "📥 Para instalar o gcloud CLI:"
        echo "   macOS: brew install google-cloud-sdk"
        echo "   Linux: https://cloud.google.com/sdk/docs/install"
        echo "   Windows: https://cloud.google.com/sdk/docs/install"
        echo ""
        exit 1
    fi
    
    log_success "gcloud CLI encontrado"
}

# Verificar autenticação
check_auth() {
    log_info "🔐 Verificando autenticação..."
    
    if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q "@"; then
        log_warning "Você não está autenticado no gcloud"
        echo ""
        echo "Execute: gcloud auth login"
        echo "Depois execute novamente este script"
        exit 1
    fi
    
    local account=$(gcloud auth list --filter=status:ACTIVE --format="value(account)")
    log_success "Autenticado como: $account"
}

# Configurar billing account
setup_billing() {
    log_info "💳 Verificando billing accounts..."
    
    # Listar billing accounts disponíveis com ID correto
    local billing_accounts=$(gcloud billing accounts list --format="table(name,displayName,open)" --filter="open:true")
    
    if [[ -z "$billing_accounts" ]] || [[ "$billing_accounts" == *"Listed 0 items"* ]]; then
        log_error "Nenhuma billing account ativa encontrada!"
        echo ""
        echo "📝 Para resolver:"
        echo "   1. Acesse: https://console.cloud.google.com/billing"
        echo "   2. Crie ou ative uma billing account"
        echo "   3. Execute novamente este script"
        echo ""
        echo "💡 Dica: Google oferece $300 de crédito grátis para novos usuários"
        exit 1
    fi
    
    echo ""
    echo "📋 Billing accounts disponíveis:"
    echo "$billing_accounts"
    echo ""
    
    # Pegar apenas o ID da billing account (primeira linha de dados)
    local billing_account_id=$(gcloud billing accounts list --format="value(name)" --filter="open:true" | head -n1)
    
    if [[ -z "$billing_account_id" ]]; then
        log_error "Não foi possível obter ID da billing account"
        exit 1
    fi
    
    log_success "Billing account ID: $billing_account_id"
    echo "$billing_account_id"
}

# Criar projetos GCP
create_projects() {
    log_info "🏗️ Criando projetos GCP..."
    
    # Obter billing account
    local billing_account_id=$(setup_billing)
    
    local projects=(
        "direito-lux-dev"
        "direito-lux-staging" 
        "direito-lux-prod"
    )
    
    for project in "${projects[@]}"; do
        log_info "Criando projeto: $project"
        
        # Criar projeto (pode falhar se já existe)
        if gcloud projects create $project --name="Direito Lux - ${project##*-}" 2>/dev/null; then
            log_success "Projeto $project criado"
            
            # Vincular billing account
            log_info "Vinculando billing account ao $project..."
            if gcloud billing projects link $project --billing-account="$billing_account_id"; then
                log_success "Billing configurado para $project"
            else
                log_warning "Falha ao configurar billing para $project - continuando..."
                log_info "Configure manualmente: https://console.cloud.google.com/billing"
            fi
        else
            log_warning "Projeto $project já existe"
            
            # Verificar se billing está configurado
            local current_billing=$(gcloud billing projects describe $project --format="value(billingAccountName)" 2>/dev/null)
            if [[ -z "$current_billing" ]]; then
                log_info "Configurando billing para projeto existente: $project"
                if gcloud billing projects link $project --billing-account="$billing_account_id" 2>/dev/null; then
                    log_success "Billing configurado para $project"
                else
                    log_warning "Falha ao configurar billing para $project - será configurado manualmente"
                    log_info "Execute: ./scripts/diagnose-billing.sh para mais detalhes"
                fi
            else
                log_success "Billing já configurado para $project: ${current_billing##*/}"
            fi
        fi
        
        # Aguardar um momento para o billing ser processado
        sleep 5
        
        # Habilitar APIs necessárias
        log_info "Habilitando APIs para $project..."
        gcloud services enable --project=$project \
            container.googleapis.com \
            sqladmin.googleapis.com \
            redis.googleapis.com \
            secretmanager.googleapis.com \
            monitoring.googleapis.com \
            logging.googleapis.com \
            cloudresourcemanager.googleapis.com \
            iam.googleapis.com \
            compute.googleapis.com \
            artifactregistry.googleapis.com \
            containerregistry.googleapis.com \
            dns.googleapis.com
        
        log_success "APIs habilitadas para $project"
    done
}

# Criar service account para Terraform
create_terraform_sa() {
    log_info "🔧 Criando Service Account para Terraform..."
    
    local project="direito-lux-dev"
    local sa_name="terraform-sa"
    local sa_email="$sa_name@$project.iam.gserviceaccount.com"
    
    # Criar service account
    if gcloud iam service-accounts create $sa_name \
        --project=$project \
        --description="Service Account para Terraform" \
        --display-name="Terraform SA" 2>/dev/null; then
        log_success "Service Account Terraform criado"
    else
        log_warning "Service Account Terraform já existe"
    fi
    
    # Dar permissões necessárias
    local roles=(
        "roles/editor"
        "roles/iam.serviceAccountAdmin" 
        "roles/resourcemanager.projectIamAdmin"
        "roles/storage.admin"
    )
    
    for role in "${roles[@]}"; do
        gcloud projects add-iam-policy-binding $project \
            --member="serviceAccount:$sa_email" \
            --role="$role" \
            --quiet
    done
    
    # Criar chave JSON
    log_info "Gerando chave JSON para Terraform..."
    gcloud iam service-accounts keys create terraform-key.json \
        --iam-account=$sa_email \
        --project=$project
    
    log_success "Chave Terraform criada: terraform-key.json"
}

# Criar service accounts para cada ambiente
create_env_service_accounts() {
    log_info "🌍 Criando Service Accounts para ambientes..."
    
    local envs=("dev" "staging" "prod")
    local projects=("direito-lux-dev" "direito-lux-staging" "direito-lux-prod")
    
    for i in "${!envs[@]}"; do
        local env="${envs[$i]}"
        local project="${projects[$i]}"
        local sa_name="github-actions-$env"
        local sa_email="$sa_name@$project.iam.gserviceaccount.com"
        
        log_info "Criando SA para ambiente: $env"
        
        # Criar service account
        if gcloud iam service-accounts create $sa_name \
            --project=$project \
            --description="GitHub Actions $env" \
            --display-name="GitHub Actions $env" 2>/dev/null; then
            log_success "Service Account $env criado"
        else
            log_warning "Service Account $env já existe"
        fi
        
        # Permissões para GKE e deploy
        local roles=(
            "roles/container.developer"
            "roles/storage.admin"
            "roles/cloudsql.client"
            "roles/redis.editor"
        )
        
        for role in "${roles[@]}"; do
            gcloud projects add-iam-policy-binding $project \
                --member="serviceAccount:$sa_email" \
                --role="$role" \
                --quiet
        done
        
        # Criar chave
        log_info "Gerando chave JSON para $env..."
        gcloud iam service-accounts keys create "github-$env-key.json" \
            --iam-account=$sa_email \
            --project=$project
        
        log_success "Chave $env criada: github-$env-key.json"
    done
}

# Exibir chaves para GitHub Secrets
show_github_secrets() {
    log_info "🔑 Preparando conteúdo para GitHub Secrets..."
    
    echo ""
    echo "════════════════════════════════════════════════════════════════"
    echo "📋 COPIE ESTES VALORES PARA O GITHUB SECRETS"
    echo "════════════════════════════════════════════════════════════════"
    echo ""
    
    # Verificar se arquivos existem
    local keys=("terraform-key.json" "github-dev-key.json" "github-staging-key.json" "github-prod-key.json")
    
    for key in "${keys[@]}"; do
        if [[ -f "$key" ]]; then
            local secret_name=""
            case "$key" in
                "terraform-key.json") secret_name="GCP_SA_KEY_TERRAFORM" ;;
                "github-dev-key.json") secret_name="GCP_SA_KEY_DEV" ;;
                "github-staging-key.json") secret_name="GCP_SA_KEY_STAGING" ;;
                "github-prod-key.json") secret_name="GCP_SA_KEY_PROD" ;;
            esac
            
            echo "🔐 Secret Name: $secret_name"
            echo "📄 Value:"
            echo "----------------------------------------"
            cat "$key" | tr -d '\n'
            echo ""
            echo "----------------------------------------"
            echo ""
        else
            log_warning "Arquivo $key não encontrado"
        fi
    done
    
    echo "📌 Project IDs:"
    echo "GCP_PROJECT_ID_DEV = direito-lux-dev"
    echo "GCP_PROJECT_ID_STAGING = direito-lux-staging" 
    echo "GCP_PROJECT_ID_PROD = direito-lux-prod"
    echo ""
    
    echo "🔗 GitHub Repository Settings:"
    echo "   1. Vá para: https://github.com/opiagile/direito-lux"
    echo "   2. Settings → Secrets and variables → Actions"
    echo "   3. New repository secret"
    echo "   4. Cole os valores acima"
    echo ""
}

# Limpar chaves locais
cleanup_keys() {
    log_warning "🧹 Removendo chaves locais por segurança..."
    
    local keys=("terraform-key.json" "github-dev-key.json" "github-staging-key.json" "github-prod-key.json")
    
    for key in "${keys[@]}"; do
        if [[ -f "$key" ]]; then
            rm -f "$key"
            log_success "Removido: $key"
        fi
    done
    
    echo ""
    log_info "🔒 Chaves removidas do computador local"
    log_info "📝 Certifique-se de ter copiado tudo para o GitHub!"
}

# Menu principal
main() {
    echo ""
    echo "🔐 Direito Lux - GitHub Secrets Setup"
    echo "====================================="
    echo ""
    
    if [[ $# -eq 0 ]]; then
        echo "Uso: $0 [comando]"
        echo ""
        echo "Comandos disponíveis:"
        echo "  setup      - Setup completo (recomendado)"
        echo "  check      - Verificar dependências"
        echo "  billing    - Verificar billing accounts"
        echo "  projects   - Criar projetos GCP"
        echo "  terraform  - Criar SA Terraform"
        echo "  envs       - Criar SAs dos ambientes"
        echo "  show       - Mostrar secrets para GitHub"
        echo "  cleanup    - Remover chaves locais"
        echo ""
        exit 1
    fi
    
    case $1 in
        "setup")
            check_gcloud
            check_auth
            create_projects
            create_terraform_sa
            create_env_service_accounts
            show_github_secrets
            echo ""
            echo "⚠️  IMPORTANTE: Copie os secrets para o GitHub antes de executar cleanup!"
            echo "Execute: $0 cleanup (depois de configurar GitHub)"
            ;;
        "check")
            check_gcloud
            check_auth
            ;;
        "billing")
            check_gcloud
            check_auth
            setup_billing
            ;;
        "projects")
            create_projects
            ;;
        "terraform")
            create_terraform_sa
            ;;
        "envs")
            create_env_service_accounts
            ;;
        "show")
            show_github_secrets
            ;;
        "cleanup")
            cleanup_keys
            ;;
        *)
            log_error "Comando desconhecido: $1"
            exit 1
            ;;
    esac
}

# Executar função principal
main "$@"