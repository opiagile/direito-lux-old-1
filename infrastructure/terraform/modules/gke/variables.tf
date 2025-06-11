# Variables já definidas no main.tf
# Arquivo mantido para compatibilidade com padrões Terraform

variable "network" {
  description = "VPC network for the cluster"
  type        = string
  default     = "default"
}

variable "subnetwork" {
  description = "VPC subnetwork for the cluster"
  type        = string
  default     = "default"
}

variable "master_ipv4_cidr_block" {
  description = "CIDR block for the master network"
  type        = string
  default     = "172.16.0.0/28"
}

variable "pods_range_name" {
  description = "Name of the secondary range for pods"
  type        = string
  default     = "pods"
}

variable "services_range_name" {
  description = "Name of the secondary range for services"
  type        = string
  default     = "services"
}