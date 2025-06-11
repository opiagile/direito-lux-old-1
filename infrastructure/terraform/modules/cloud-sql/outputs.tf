output "instance_name" {
  description = "Cloud SQL instance name"
  value       = google_sql_database_instance.postgres.name
}

output "instance_connection_name" {
  description = "Cloud SQL instance connection name"
  value       = google_sql_database_instance.postgres.connection_name
}

output "private_ip_address" {
  description = "Private IP address of the Cloud SQL instance"
  value       = google_sql_database_instance.postgres.private_ip_address
}

output "public_ip_address" {
  description = "Public IP address of the Cloud SQL instance"
  value       = google_sql_database_instance.postgres.first_ip_address
}

output "database_name" {
  description = "Database name"
  value       = google_sql_database.direito_lux.name
}

output "app_user" {
  description = "Application database user"
  value       = google_sql_user.app_user.name
}

output "readonly_user" {
  description = "Readonly database user"
  value       = google_sql_user.readonly_user.name
}

output "server_ca_cert" {
  description = "Server CA certificate"
  value       = google_sql_database_instance.postgres.server_ca_cert.0.cert
  sensitive   = true
}

output "client_cert" {
  description = "Client certificate"
  value       = google_sql_ssl_cert.client_cert.cert
  sensitive   = true
}

output "client_key" {
  description = "Client private key"
  value       = google_sql_ssl_cert.client_cert.private_key
  sensitive   = true
}