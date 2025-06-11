# GKE Cluster Module
# Cria cluster Kubernetes otimizado para Direito Lux

variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "region" {
  description = "GCP Region"
  type        = string
}

variable "environment" {
  description = "Environment name"
  type        = string
}

variable "cluster_name" {
  description = "GKE Cluster name"
  type        = string
}

variable "node_pools" {
  description = "Node pools configuration"
  type = map(object({
    machine_type = string
    min_nodes    = number
    max_nodes    = number
    disk_size_gb = number
  }))
}

variable "enable_autopilot" {
  description = "Enable GKE Autopilot"
  type        = bool
  default     = false
}

variable "enable_private_cluster" {
  description = "Enable private cluster"
  type        = bool
  default     = true
}

# VPC Network
resource "google_compute_network" "vpc" {
  name                    = "${var.cluster_name}-vpc"
  auto_create_subnetworks = false
  routing_mode           = "GLOBAL"
}

resource "google_compute_subnetwork" "subnet" {
  name          = "${var.cluster_name}-subnet"
  ip_cidr_range = "10.0.0.0/24"
  region        = var.region
  network       = google_compute_network.vpc.name
  
  secondary_ip_range {
    range_name    = "pods"
    ip_cidr_range = "10.1.0.0/16"
  }
  
  secondary_ip_range {
    range_name    = "services"
    ip_cidr_range = "10.2.0.0/16"
  }
}

# Service Account for GKE nodes
resource "google_service_account" "gke_nodes" {
  account_id   = "${var.cluster_name}-nodes"
  display_name = "GKE Nodes Service Account"
}

resource "google_project_iam_member" "gke_nodes_roles" {
  for_each = toset([
    "roles/logging.logWriter",
    "roles/monitoring.metricWriter",
    "roles/monitoring.viewer",
    "roles/storage.objectViewer"
  ])
  
  project = var.project_id
  role    = each.value
  member  = "serviceAccount:${google_service_account.gke_nodes.email}"
}

# GKE Cluster
resource "google_container_cluster" "primary" {
  name     = var.cluster_name
  location = var.region
  
  # Remove default node pool
  remove_default_node_pool = true
  initial_node_count       = 1
  
  # Network configuration
  network    = google_compute_network.vpc.name
  subnetwork = google_compute_subnetwork.subnet.name
  
  # Private cluster configuration
  dynamic "private_cluster_config" {
    for_each = var.enable_private_cluster ? [1] : []
    content {
      enable_private_nodes    = true
      enable_private_endpoint = false
      master_ipv4_cidr_block = "172.16.0.0/28"
    }
  }
  
  # IP allocation policy
  ip_allocation_policy {
    cluster_secondary_range_name  = "pods"
    services_secondary_range_name = "services"
  }
  
  # Workload Identity
  workload_identity_config {
    workload_pool = "${var.project_id}.svc.id.goog"
  }
  
  # Addons
  addons_config {
    horizontal_pod_autoscaling {
      disabled = false
    }
    
    http_load_balancing {
      disabled = false
    }
    
    network_policy_config {
      disabled = false
    }
    
    istio_config {
      disabled = false
      auth     = "AUTH_MUTUAL_TLS"
    }
  }
  
  # Network policy
  network_policy {
    enabled = true
  }
  
  # Maintenance policy
  maintenance_policy {
    daily_maintenance_window {
      start_time = "03:00"
    }
  }
  
  # Logging and monitoring
  logging_service    = "logging.googleapis.com/kubernetes"
  monitoring_service = "monitoring.googleapis.com/kubernetes"
  
  # Security
  enable_shielded_nodes = true
  
  # Resource labels
  resource_labels = {
    environment = var.environment
    project     = "direito-lux"
    managed-by  = "terraform"
  }
}

# Node pools
resource "google_container_node_pool" "pools" {
  for_each = var.node_pools
  
  name       = each.key
  location   = var.region
  cluster    = google_container_cluster.primary.name
  node_count = each.value.min_nodes
  
  # Auto scaling
  autoscaling {
    min_node_count = each.value.min_nodes
    max_node_count = each.value.max_nodes
  }
  
  # Node configuration
  node_config {
    preemptible  = var.environment != "prod"
    machine_type = each.value.machine_type
    disk_size_gb = each.value.disk_size_gb
    disk_type    = "pd-ssd"
    
    # Service account
    service_account = google_service_account.gke_nodes.email
    
    # OAuth scopes
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]
    
    # Workload Identity
    workload_metadata_config {
      mode = "GKE_METADATA"
    }
    
    # Shielded nodes
    shielded_instance_config {
      enable_secure_boot          = true
      enable_integrity_monitoring = true
    }
    
    # Labels
    labels = {
      environment = var.environment
      pool        = each.key
    }
    
    # Taints for specific workloads
    dynamic "taint" {
      for_each = each.key == "gpu" ? [1] : []
      content {
        key    = "nvidia.com/gpu"
        value  = "present"
        effect = "NO_SCHEDULE"
      }
    }
  }
  
  # Upgrade settings
  upgrade_settings {
    max_surge       = 1
    max_unavailable = 0
  }
  
  # Management
  management {
    auto_repair  = true
    auto_upgrade = true
  }
}

# Outputs
output "cluster_name" {
  description = "GKE cluster name"
  value       = google_container_cluster.primary.name
}

output "endpoint" {
  description = "GKE cluster endpoint"
  value       = google_container_cluster.primary.endpoint
  sensitive   = true
}

output "ca_certificate" {
  description = "GKE cluster CA certificate"
  value       = google_container_cluster.primary.master_auth.0.cluster_ca_certificate
  sensitive   = true
}

output "network" {
  description = "VPC network name"
  value       = google_compute_network.vpc.name
}

output "subnet" {
  description = "Subnet name"
  value       = google_compute_subnetwork.subnet.name
}