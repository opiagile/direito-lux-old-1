# Direito Lux - Ambiente Development
# Terraform configuration for dev environment

terraform {
  required_version = ">= 1.5.0"
  
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.23"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.11"
    }
  }
  
  # Remote state em GCS
  backend "gcs" {
    bucket = "direito-lux-terraform-state"
    prefix = "dev/terraform.tfstate"
  }
}

# Variáveis do ambiente
variable "project_id" {
  description = "GCP Project ID"
  type        = string
  default     = "direito-lux-dev"
}

variable "region" {
  description = "GCP Region"
  type        = string
  default     = "us-central1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

# Provider configuration
provider "google" {
  project = var.project_id
  region  = var.region
}

# GKE Cluster
module "gke_cluster" {
  source = "../../modules/gke"
  
  project_id   = var.project_id
  region       = var.region
  environment  = var.environment
  cluster_name = "direito-lux-${var.environment}"
  
  # Dev usa cluster menor
  node_pools = {
    default = {
      machine_type = "e2-standard-2"
      min_nodes    = 1
      max_nodes    = 3
      disk_size_gb = 50
    }
  }
  
  # Features
  enable_autopilot = false
  enable_private_cluster = false  # Dev pode ser público
}

# VPC Network
resource "google_compute_network" "vpc" {
  name                    = "${var.environment}-vpc"
  auto_create_subnetworks = false
  project                 = var.project_id
}

resource "google_compute_subnetwork" "subnet" {
  name          = "${var.environment}-subnet"
  ip_cidr_range = "10.0.0.0/16"
  region        = var.region
  network       = google_compute_network.vpc.id
  project       = var.project_id

  secondary_ip_range {
    range_name    = "pods"
    ip_cidr_range = "10.1.0.0/16"
  }

  secondary_ip_range {
    range_name    = "services"
    ip_cidr_range = "10.2.0.0/16"
  }
}

# Cloud SQL (PostgreSQL)
module "cloud_sql" {
  source = "../../modules/cloud-sql"
  
  project_id   = var.project_id
  region       = var.region
  environment  = var.environment
  vpc_network  = google_compute_network.vpc.id
  
  # Dev usa instância menor
  db_tier               = "db-f1-micro"
  disk_size            = 20
  disk_autoresize_limit = 100
  
  # Senhas (usar Secret Manager em produção)
  app_db_password      = "DireitoLux2024Dev!"
  readonly_db_password = "ReadOnly2024Dev!"
}

# Redis (Memorystore)
module "redis" {
  source = "../../modules/memorystore"
  
  project_id   = var.project_id
  region       = var.region
  environment  = var.environment
  
  # Dev usa Redis básico
  tier          = "BASIC"
  memory_size_gb = 1
  redis_version  = "REDIS_7_0"
}

# Load Balancer
module "load_balancer" {
  source = "../../modules/load-balancer"
  
  project_id   = var.project_id
  region       = var.region
  environment  = var.environment
  
  # SSL Certificate
  managed_ssl_certificate_domains = [
    "dev.direito-lux.com.br"
  ]
}

# IAM & Service Accounts
module "iam" {
  source = "../../modules/iam"
  
  project_id   = var.project_id
  environment  = var.environment
  
  service_accounts = {
    gke_nodes = {
      display_name = "GKE Nodes SA"
      roles = [
        "roles/logging.logWriter",
        "roles/monitoring.metricWriter",
        "roles/monitoring.viewer"
      ]
    }
    workload_identity = {
      display_name = "Workload Identity SA"
      roles = [
        "roles/cloudsql.client",
        "roles/redis.editor",
        "roles/secretmanager.secretAccessor"
      ]
    }
  }
}

# Secrets Manager
module "secrets" {
  source = "../../modules/secrets"
  
  project_id   = var.project_id
  environment  = var.environment
  
  secrets = {
    db_password = {
      description = "PostgreSQL password"
    }
    redis_password = {
      description = "Redis password"
    }
    keycloak_admin_password = {
      description = "Keycloak admin password"
    }
    openai_api_key = {
      description = "OpenAI API key"
    }
  }
}

# Monitoring & Logging
module "monitoring" {
  source = "../../modules/monitoring"
  
  project_id   = var.project_id
  environment  = var.environment
  
  # Dev tem alertas mais relaxados
  alert_config = {
    cpu_threshold    = 80
    memory_threshold = 85
    disk_threshold   = 90
  }
}

# Outputs
output "gke_cluster_name" {
  value = module.gke_cluster.cluster_name
}

output "gke_cluster_endpoint" {
  value     = module.gke_cluster.endpoint
  sensitive = true
}

output "cloud_sql_instance_name" {
  value = module.cloud_sql.instance_name
}

output "cloud_sql_connection_name" {
  value = module.cloud_sql.instance_connection_name
}

output "cloud_sql_private_ip" {
  value = module.cloud_sql.private_ip_address
}

output "redis_host" {
  value = module.redis.host
}

output "load_balancer_ip" {
  value = module.load_balancer.external_ip
}