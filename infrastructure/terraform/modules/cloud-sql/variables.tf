variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, staging, production)"
  type        = string
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "us-central1"
}

variable "vpc_network" {
  description = "VPC network for private IP"
  type        = string
}

variable "db_tier" {
  description = "Database tier"
  type        = string
  default     = "db-f1-micro"
}

variable "disk_size" {
  description = "Initial disk size in GB"
  type        = number
  default     = 20
}

variable "disk_autoresize_limit" {
  description = "Maximum disk size in GB"
  type        = number
  default     = 100
}

variable "app_db_password" {
  description = "Password for application database user"
  type        = string
  sensitive   = true
}

variable "readonly_db_password" {
  description = "Password for readonly database user"
  type        = string
  sensitive   = true
}