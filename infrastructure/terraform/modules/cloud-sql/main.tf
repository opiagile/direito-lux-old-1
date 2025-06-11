# Cloud SQL PostgreSQL Instance
resource "google_sql_database_instance" "postgres" {
  name             = "${var.environment}-postgres-${random_id.db_name_suffix.hex}"
  database_version = "POSTGRES_16"
  region          = var.region
  project         = var.project_id

  settings {
    tier                        = var.db_tier
    availability_type          = var.environment == "production" ? "REGIONAL" : "ZONAL"
    disk_type                  = "PD_SSD"
    disk_size                  = var.disk_size
    disk_autoresize           = true
    disk_autoresize_limit     = var.disk_autoresize_limit

    backup_configuration {
      enabled                        = true
      start_time                    = "03:00"
      location                      = var.region
      point_in_time_recovery_enabled = true
      transaction_log_retention_days = 7
      backup_retention_settings {
        retained_backups = var.environment == "production" ? 30 : 7
        retention_unit   = "COUNT"
      }
    }

    ip_configuration {
      ipv4_enabled       = false
      private_network    = var.vpc_network
      enable_private_path_for_google_cloud_services = true
    }

    database_flags {
      name  = "log_statement"
      value = var.environment == "production" ? "ddl" : "all"
    }

    database_flags {
      name  = "log_min_duration_statement"
      value = "1000"
    }

    database_flags {
      name  = "shared_preload_libraries"
      value = "pg_stat_statements"
    }

    maintenance_window {
      day          = 7  # Sunday
      hour         = 3  # 3 AM
      update_track = "stable"
    }

    insights_config {
      query_insights_enabled  = true
      query_string_length    = 1024
      record_application_tags = true
      record_client_address  = true
    }

    deletion_protection_enabled = var.environment == "production"
  }

  deletion_protection = var.environment == "production"

  depends_on = [google_service_networking_connection.private_vpc_connection]
}

# Random suffix for database instance name
resource "random_id" "db_name_suffix" {
  byte_length = 4
}

# Database
resource "google_sql_database" "direito_lux" {
  name     = "direito_lux"
  instance = google_sql_database_instance.postgres.name
  charset  = "UTF8"
  collation = "en_US.UTF8"
}

# Database users
resource "google_sql_user" "app_user" {
  name     = "direito_lux_app"
  instance = google_sql_database_instance.postgres.name
  password = var.app_db_password
}

resource "google_sql_user" "readonly_user" {
  name     = "direito_lux_readonly"
  instance = google_sql_database_instance.postgres.name
  password = var.readonly_db_password
}

# Private service networking for Cloud SQL
resource "google_compute_global_address" "private_ip_address" {
  name          = "${var.environment}-private-ip-address"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = var.vpc_network
  project       = var.project_id
}

resource "google_service_networking_connection" "private_vpc_connection" {
  network                 = var.vpc_network
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_address.name]
}

# Cloud SQL SSL certificate
resource "google_sql_ssl_cert" "client_cert" {
  common_name = "${var.environment}-direito-lux-cert"
  instance    = google_sql_database_instance.postgres.name
  project     = var.project_id
}