# Terraform outputs for Agro Mas Backend

output "project_id" {
  description = "The GCP project ID"
  value       = var.project_id
}

output "region" {
  description = "The GCP region"
  value       = var.region
}

output "environment" {
  description = "The deployment environment"
  value       = var.environment
}

# Database outputs
output "database_instance_name" {
  description = "Cloud SQL instance name"
  value       = google_sql_database_instance.agro_mas_db.name
}

output "database_connection_name" {
  description = "Cloud SQL connection name"
  value       = google_sql_database_instance.agro_mas_db.connection_name
}

output "database_private_ip" {
  description = "Cloud SQL private IP address"
  value       = google_sql_database_instance.agro_mas_db.private_ip_address
  sensitive   = true
}

output "database_name" {
  description = "Database name"
  value       = google_sql_database.agro_mas_database.name
}

output "database_user" {
  description = "Database user"
  value       = google_sql_user.agro_mas_user.name
}

# Storage outputs
output "storage_bucket_name" {
  description = "Cloud Storage bucket name"
  value       = google_storage_bucket.agro_mas_storage.name
}

output "storage_bucket_url" {
  description = "Cloud Storage bucket URL"
  value       = google_storage_bucket.agro_mas_storage.url
}

# Network outputs
output "vpc_network_name" {
  description = "VPC network name"
  value       = google_compute_network.agro_mas_vpc.name
}

output "vpc_subnet_name" {
  description = "VPC subnet name"
  value       = google_compute_subnetwork.agro_mas_subnet.name
}

output "vpc_connector_name" {
  description = "VPC connector name"
  value       = google_vpc_access_connector.agro_mas_connector.name
}

# Service account outputs
output "service_account_email" {
  description = "Service account email"
  value       = google_service_account.agro_mas_backend.email
}

# Secret Manager outputs
output "jwt_secret_name" {
  description = "JWT secret name in Secret Manager"
  value       = google_secret_manager_secret.jwt_secret.secret_id
}

output "db_password_secret_name" {
  description = "Database password secret name in Secret Manager"
  value       = google_secret_manager_secret.db_password.secret_id
}

# Cloud Build outputs
output "cloud_build_trigger_name" {
  description = "Cloud Build trigger name"
  value       = google_cloudbuild_trigger.agro_mas_trigger.name
}

output "cloud_build_trigger_id" {
  description = "Cloud Build trigger ID"
  value       = google_cloudbuild_trigger.agro_mas_trigger.trigger_id
}

# Monitoring outputs
output "notification_channel_name" {
  description = "Monitoring notification channel name"
  value       = google_monitoring_notification_channel.email.name
}

# URLs and endpoints
output "cloud_run_url" {
  description = "Cloud Run service URL (will be available after deployment)"
  value       = "https://agro-mas-backend-${random_id.suffix.hex}-${var.region}.a.run.app"
}

output "health_check_url" {
  description = "Health check endpoint URL"
  value       = "https://agro-mas-backend-${random_id.suffix.hex}-${var.region}.a.run.app/health"
}

output "api_base_url" {
  description = "API base URL"
  value       = "https://agro-mas-backend-${random_id.suffix.hex}-${var.region}.a.run.app/api/v1"
}

# Random suffix for unique resource names
resource "random_id" "suffix" {
  byte_length = 4
}

# Configuration summary
output "deployment_summary" {
  description = "Summary of the deployment configuration"
  value = {
    environment     = var.environment
    region         = var.region
    database_tier  = local.config.db_tier
    database_size  = "${local.config.db_disk_size}GB"
    min_instances  = local.config.min_instances
    max_instances  = local.config.max_instances
    memory         = local.config.memory
    cpu            = local.config.cpu
    backup_enabled = local.config.backup_enabled
  }
}

# Cost estimation (rough)
output "estimated_monthly_cost" {
  description = "Rough estimation of monthly costs in USD"
  value = {
    database = var.environment == "production" ? "$50-100" : "$15-30"
    cloud_run = var.environment == "production" ? "$20-50" : "$5-15"
    storage = "$5-20"
    networking = "$5-10"
    monitoring = "Free tier"
    total_estimate = var.environment == "production" ? "$80-180" : "$25-75"
    note = "Actual costs depend on usage patterns and data transfer"
  }
}

# Important commands
output "useful_commands" {
  description = "Useful commands for managing the deployment"
  value = {
    connect_to_db = "gcloud sql connect ${google_sql_database_instance.agro_mas_db.name} --user=${google_sql_user.agro_mas_user.name}"
    view_logs = "gcloud logging read 'resource.type=\"cloud_run_revision\" AND resource.labels.service_name=\"agro-mas-backend\"' --limit 50 --format json"
    deploy_manually = "gcloud run deploy agro-mas-backend --image gcr.io/${var.project_id}/agro-mas-backend:latest --region ${var.region}"
    view_metrics = "gcloud monitoring metrics list --filter='resource.type=\"cloud_run_revision\"'"
  }
}