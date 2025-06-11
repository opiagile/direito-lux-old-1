#!/bin/bash

# 🔐 Setup GCP Service Accounts para Direito Lux
# Este script cria todas as service accounts necessárias e gera as chaves JSON

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Função para log
log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Verificar se gcloud está instalado
if ! command -v gcloud &> /dev/null; then
    error "gcloud CLI não está instalado. Instale: https://cloud.google.com/sdk/docs/install"
    exit 1
fi

# Verificar se está autenticado
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q .; then
    error "Não está autenticado no gcloud. Execute: gcloud auth login"
    exit 1
fi

# Verificar se jq está instalado
if ! command -v jq &> /dev/null; then
    error "jq não está instalado. Instale: brew install jq"
    exit 1
fi

# Configurações dos projetos
PROJECTS=(
    "direito-lux-dev:dev"
    "direito-lux-staging:staging" 
    "direito-lux-prod:prod"
)

# Roles necessárias
ROLES=(
    "roles/compute.admin"
    "roles/container.admin"
    "roles/cloudsql.admin"
    "roles/redis.admin"
    "roles/secretmanager.admin"
    "roles/storage.admin"
    "roles/iam.serviceAccountAdmin"
    "roles/resourcemanager.projectIamAdmin"
    "roles/serviceusage.serviceUsageAdmin"
    "roles/dns.admin"
    "roles/logging.admin"
    "roles/monitoring.admin"
)

# Função para criar service account
create_service_account() {
    local project=$1
    local sa_name=$2
    local sa_display_name=$3
    local env=$4
    
    log "Criando service account: ${sa_name} no projeto: ${project}"
    
    # Criar service account se não existir
    if ! gcloud iam service-accounts describe "${sa_name}@${project}.iam.gserviceaccount.com" \
         --project="${project}" &>/dev/null; then
        gcloud iam service-accounts create "${sa_name}" \
            --display-name="${sa_display_name}" \
            --description="Service account para ${env} - ${sa_display_name}" \
            --project="${project}"
        log "Service account ${sa_name} criada com sucesso"
    else
        warn "Service account ${sa_name} já existe"
    fi
    
    # Adicionar roles
    for role in "${ROLES[@]}"; do
        log "Adicionando role: ${role}"
        gcloud projects add-iam-policy-binding "${project}" \
            --member="serviceAccount:${sa_name}@${project}.iam.gserviceaccount.com" \
            --role="${role}" \
            --quiet
    done
}

# Função para gerar chave JSON
generate_key() {
    local project=$1
    local sa_name=$2
    local env=$3
    
    local key_file="keys/${sa_name}-${env}-key.json"
    
    log "Gerando chave JSON para: ${sa_name}@${project}.iam.gserviceaccount.com"
    
    # Criar diretório keys se não existir
    mkdir -p keys
    
    # Gerar nova chave
    gcloud iam service-accounts keys create "${key_file}" \
        --iam-account="${sa_name}@${project}.iam.gserviceaccount.com" \
        --project="${project}"
    
    log "Chave salva em: ${key_file}"
    
    # Mostrar conteúdo formatado para GitHub Secrets
    echo -e "\n${BLUE}=== GitHub Secret: GCP_SA_KEY_$(echo ${env} | tr '[:lower:]' '[:upper:]') ===${NC}"
    jq -c . < "${key_file}"
    echo ""
}

# Função para habilitar APIs
enable_apis() {
    local project=$1
    
    log "Habilitando APIs necessárias no projeto: ${project}"
    gcloud services enable \
        compute.googleapis.com \
        container.googleapis.com \
        sqladmin.googleapis.com \
        redis.googleapis.com \
        secretmanager.googleapis.com \
        storage.googleapis.com \
        iam.googleapis.com \
        cloudresourcemanager.googleapis.com \
        serviceusage.googleapis.com \
        dns.googleapis.com \
        logging.googleapis.com \
        monitoring.googleapis.com \
        --project="${project}" \
        --quiet
}

# Função principal
main() {
    log "🚀 Iniciando setup das Service Accounts GCP para Direito Lux"
    
    # Verificar se os projetos existem e habilitar APIs
    for project_env in "${PROJECTS[@]}"; do
        IFS=':' read -r project env <<< "$project_env"
        
        log "Verificando projeto: ${project}"
        
        if ! gcloud projects describe "${project}" &>/dev/null; then
            error "Projeto ${project} não existe ou você não tem acesso"
            exit 1
        fi
        
        enable_apis "${project}"
    done
    
    # Criar service accounts e chaves
    for project_env in "${PROJECTS[@]}"; do
        IFS=':' read -r project env <<< "$project_env"
        
        log "=== Configurando ambiente: ${env} (${project}) ==="
        
        # Terraform SA (apenas para dev, usado para gerenciar infraestrutura)
        if [[ "$env" == "dev" ]]; then
            create_service_account "${project}" "terraform-sa" "Terraform Service Account" "${env}"
            generate_key "${project}" "terraform-sa" "terraform"
        fi
        
        # GitHub Actions SA (para cada ambiente)
        local gh_sa_name="github-actions-${env}"
        create_service_account "${project}" "${gh_sa_name}" "GitHub Actions $(echo ${env:0:1} | tr '[:lower:]' '[:upper:]')$(echo ${env:1}) Service Account" "${env}"
        generate_key "${project}" "${gh_sa_name}" "${env}"
    done
    
    # Resumo final
    echo -e "\n${GREEN}🎉 Setup concluído com sucesso!${NC}"
    echo -e "\n${BLUE}📋 Próximos passos:${NC}"
    echo "1. Vá para: https://github.com/opiagile/direito-lux/settings/secrets/actions"
    echo "2. Adicione os seguintes Repository Secrets:"
    echo "   - GCP_SA_KEY_TERRAFORM (conteúdo do terraform-sa-terraform-key.json)"
    echo "   - GCP_SA_KEY_DEV (conteúdo do github-actions-dev-dev-key.json)"
    echo "   - GCP_SA_KEY_STAGING (conteúdo do github-actions-staging-staging-key.json)"
    echo "   - GCP_SA_KEY_PROD (conteúdo do github-actions-prod-prod-key.json)"
    echo ""
    echo "3. Adicione também as variáveis de repositório:"
    echo "   - ENABLE_GCP_DEPLOY = true"
    echo ""
    echo -e "${YELLOW}⚠️  IMPORTANTE:${NC}"
    echo "- Delete os arquivos JSON após configurar os secrets"
    echo "- Não commite as chaves no Git"
    echo "- Mantenha as chaves seguras"
    
    # Mostrar localização dos arquivos
    echo -e "\n${BLUE}📁 Arquivos gerados:${NC}"
    ls -la keys/*.json 2>/dev/null || echo "Nenhum arquivo encontrado"
}

# Função para limpeza
cleanup() {
    read -p "Deseja deletar os arquivos de chave JSON? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf keys/
        log "Arquivos de chave deletados"
    fi
}

# Executar script
main "$@"

# Perguntar sobre limpeza
echo ""
cleanup