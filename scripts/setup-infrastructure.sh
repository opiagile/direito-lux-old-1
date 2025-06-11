#!/bin/bash

# Direito Lux - Infrastructure Setup Script
# Configura toda a infraestrutura Cloud + CI/CD

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

# Verificar dependências
check_dependencies() {
    log_info "🔍 Verificando dependências..."
    
    local missing_deps=()
    
    # Verificar ferramentas necessárias
    for tool in gcloud terraform kubectl helm argocd; do
        if ! command -v $tool &> /dev/null; then
            missing_deps+=($tool)
        fi
    done
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "Dependências faltando: ${missing_deps[*]}"
        log_info "Instale as dependências e execute novamente"
        exit 1
    fi
    
    log_success "Todas as dependências encontradas"
}

# Configurar projetos GCP
setup_gcp_projects() {
    log_info "🏗️ Configurando projetos GCP..."
    
    # Projetos para cada ambiente
    local projects=(
        "direito-lux-dev"
        "direito-lux-staging" 
        "direito-lux-prod"
    )
    
    for project in "${projects[@]}"; do
        log_info "Criando projeto: $project"
        
        # Criar projeto (pode falhar se já existe)
        gcloud projects create $project --name="Direito Lux - ${project##*-}" || true
        
        # Habilitar APIs necessárias
        gcloud services enable --project=$project \
            container.googleapis.com \
            sqladmin.googleapis.com \
            redis.googleapis.com \
            secretmanager.googleapis.com \
            monitoring.googleapis.com \
            logging.googleapis.com \
            cloudresourcemanager.googleapis.com \
            iam.googleapis.com
        
        log_success "Projeto $project configurado"
    done
}

# Configurar Terraform backend
setup_terraform_backend() {
    log_info "🗄️ Configurando Terraform backend..."
    
    local bucket_name="direito-lux-terraform-state"
    local project="direito-lux-dev"
    
    # Criar bucket para state
    gsutil mb -p $project gs://$bucket_name 2>/dev/null || true
    
    # Habilitar versionamento
    gsutil versioning set on gs://$bucket_name
    
    # Configurar lifecycle
    cat > lifecycle.json << EOF
{
  "rule": [
    {
      "action": {"type": "Delete"},
      "condition": {"age": 90, "isLive": false}
    }
  ]
}
EOF
    
    gsutil lifecycle set lifecycle.json gs://$bucket_name
    rm lifecycle.json
    
    log_success "Terraform backend configurado"
}

# Aplicar Terraform para DEV
apply_terraform_dev() {
    log_info "🚀 Aplicando Terraform para DEV..."
    
    cd infrastructure/terraform/environments/dev
    
    # Inicializar
    terraform init
    
    # Planejar
    terraform plan -var="project_id=direito-lux-dev" -out=tfplan
    
    # Aplicar
    log_warning "Aplicando infraestrutura DEV..."
    terraform apply tfplan
    
    cd ../../../../
    
    log_success "Infraestrutura DEV criada"
}

# Instalar ArgoCD
install_argocd() {
    log_info "⚙️ Instalando ArgoCD..."
    
    # Conectar ao cluster DEV
    gcloud container clusters get-credentials direito-lux-dev --region=us-central1 --project=direito-lux-dev
    
    # Criar namespace
    kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
    
    # Instalar ArgoCD
    kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
    
    # Aguardar pods ficarem prontos
    log_info "Aguardando ArgoCD ficar pronto..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=argocd-server -n argocd --timeout=300s
    
    # Obter senha inicial
    local argocd_password=$(kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d)
    
    # Aplicar configurações customizadas
    kubectl apply -f infrastructure/argocd/applications/
    
    log_success "ArgoCD instalado"
    log_info "ArgoCD UI: kubectl port-forward svc/argocd-server -n argocd 8080:443"
    log_info "Usuário: admin"
    log_info "Senha: $argocd_password"
}

# Configurar secrets no GitHub
setup_github_secrets() {
    log_info "🔐 Configurando GitHub Secrets..."
    
    log_warning "Configure manualmente os seguintes secrets no GitHub:"
    echo ""
    echo "Repository Secrets:"
    echo "  GCP_SA_KEY - Service Account key para Terraform"
    echo "  GCP_SA_KEY_DEV - Service Account key para DEV"
    echo "  GCP_SA_KEY_STAGING - Service Account key para STAGING"  
    echo "  GCP_SA_KEY_PROD - Service Account key para PROD"
    echo "  GCP_PROJECT_ID_DEV - direito-lux-dev"
    echo "  GCP_PROJECT_ID_STAGING - direito-lux-staging"
    echo "  GCP_PROJECT_ID_PROD - direito-lux-prod"
    echo "  SLACK_WEBHOOK - Webhook do Slack para notificações"
    echo "  INFRACOST_API_KEY - API key do Infracost"
    echo ""
    echo "Environment Secrets (por ambiente):"
    echo "  PROD_DB_CONNECTION - String de conexão do banco de produção"
    echo ""
}

# Configurar monitoramento inicial
setup_monitoring() {
    log_info "📊 Configurando monitoramento..."
    
    # Instalar Prometheus/Grafana via Helm
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    helm repo update
    
    # Instalar kube-prometheus-stack
    helm install monitoring prometheus-community/kube-prometheus-stack \
        --namespace monitoring \
        --create-namespace \
        --set grafana.adminPassword=admin123 \
        --set prometheus.prometheusSpec.retention=7d
    
    log_success "Monitoramento configurado"
    log_info "Grafana: kubectl port-forward svc/monitoring-grafana -n monitoring 3000:80"
    log_info "Usuário: admin / Senha: admin123"
}

# Teste de infraestrutura
test_infrastructure() {
    log_info "🧪 Testando infraestrutura..."
    
    # Verificar cluster
    kubectl cluster-info
    
    # Verificar nodes
    kubectl get nodes
    
    # Verificar ArgoCD
    kubectl get pods -n argocd
    
    # Verificar monitoring
    kubectl get pods -n monitoring
    
    log_success "Infraestrutura funcionando corretamente"
}

# Menu principal
main() {
    echo ""
    echo "🏗️  Direito Lux - Infrastructure Setup"
    echo "======================================"
    echo ""
    
    if [[ $# -eq 0 ]]; then
        echo "Uso: $0 [comando]"
        echo ""
        echo "Comandos disponíveis:"
        echo "  full        - Setup completo (recomendado)"
        echo "  check       - Verificar dependências"
        echo "  gcp         - Configurar projetos GCP"
        echo "  terraform   - Aplicar Terraform DEV"
        echo "  argocd      - Instalar ArgoCD"
        echo "  monitoring  - Configurar monitoramento"
        echo "  test        - Testar infraestrutura"
        echo "  secrets     - Instruções para GitHub Secrets"
        echo ""
        exit 1
    fi
    
    case $1 in
        "full")
            check_dependencies
            setup_gcp_projects
            setup_terraform_backend
            apply_terraform_dev
            install_argocd
            setup_monitoring
            test_infrastructure
            setup_github_secrets
            log_success "🎉 Setup completo finalizado!"
            ;;
        "check")
            check_dependencies
            ;;
        "gcp")
            setup_gcp_projects
            ;;
        "terraform")
            apply_terraform_dev
            ;;
        "argocd")
            install_argocd
            ;;
        "monitoring")
            setup_monitoring
            ;;
        "test")
            test_infrastructure
            ;;
        "secrets")
            setup_github_secrets
            ;;
        *)
            log_error "Comando desconhecido: $1"
            exit 1
            ;;
    esac
}

# Executar função principal
main "$@"