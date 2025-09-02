# Terraform configuration for Agro Mas Backend infrastructure on Google Cloud

terraform {
  required_version = ">= 1.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 5.0"
    }
  }
  
  # Store state in Google Cloud Storage
  backend "gcs" {
    bucket = "agro-mas-terraform-state"
    prefix = "backend/state"
  }
}

# Configure the Google Cloud Provider
provider "google" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}

provider "google-beta" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}

# Enable required APIs
resource "google_project_service" "apis" {
  for_each = toset([
    "cloudresourcemanager.googleapis.com",
    "cloudbuild.googleapis.com",
    "run.googleapis.com",
    "sqladmin.googleapis.com",
    "storage.googleapis.com",
    "vpcaccess.googleapis.com",
    "secretmanager.googleapis.com",
    "cloudmonitoring.googleapis.com",
    "cloudlogging.googleapis.com",
    "cloudsql.googleapis.com",
    "compute.googleapis.com",
  ])
  
  service = each.value
  
  disable_dependent_services = true
  disable_on_destroy        = false
}

# Create a VPC network for private resources
resource "google_compute_network" "agro_mas_vpc" {
  name                    = "agro-mas-vpc"
  auto_create_subnetworks = false
  
  depends_on = [google_project_service.apis]
}

# Create a subnet for the VPC
resource "google_compute_subnetwork" "agro_mas_subnet" {
  name          = "agro-mas-subnet"
  ip_cidr_range = "10.0.0.0/24"
  region        = var.region
  network       = google_compute_network.agro_mas_vpc.id
  
  private_ip_google_access = true
}

# Create VPC connector for Cloud Run to access private resources
resource "google_vpc_access_connector" "agro_mas_connector" {
  provider = google-beta
  
  name          = "agro-mas-connector"
  region        = var.region
  network       = google_compute_network.agro_mas_vpc.name
  ip_cidr_range = "10.8.0.0/28"
  min_instances = 2
  max_instances = 10
  
  depends_on = [google_project_service.apis]
}

# Cloud SQL instance for PostgreSQL
resource "google_sql_database_instance" "agro_mas_db" {
  name             = "agro-mas-db-${var.environment}"
  database_version = "POSTGRES_15"
  region           = var.region
  
  deletion_protection = var.environment == "production" ? true : false
  
  settings {
    tier              = var.db_tier
    availability_type = var.environment == "production" ? "REGIONAL" : "ZONAL"
    disk_size         = var.db_disk_size
    disk_type         = "PD_SSD"
    
    backup_configuration {
      enabled                        = true
      start_time                     = "03:00"
      location                       = var.region
      point_in_time_recovery_enabled = true
      backup_retention_settings {
        retained_backups = 30
      }
    }
    
    ip_configuration {
      ipv4_enabled    = false
      private_network = google_compute_network.agro_mas_vpc.id
      require_ssl     = true
    }
    
    database_flags {
      name  = "max_connections"
      value = "100"
    }
    
    database_flags {
      name  = "shared_preload_libraries"
      value = "postgis"
    }
    
    insights_config {
      query_insights_enabled  = true
      query_string_length     = 1024
      record_application_tags = true
      record_client_address   = true
    }
  }
  
  depends_on = [
    google_project_service.apis,
    google_compute_network.agro_mas_vpc
  ]
}

# Create database
resource "google_sql_database" "agro_mas_database" {
  name     = "agro_mas_${var.environment}"
  instance = google_sql_database_instance.agro_mas_db.name
}

# Create database user
resource "google_sql_user" "agro_mas_user" {
  name     = "agro_mas_user"
  instance = google_sql_database_instance.agro_mas_db.name
  password = var.db_password
}

# Cloud Storage bucket for file uploads
resource "google_storage_bucket" "agro_mas_storage" {
  name          = "agro-mas-storage-${var.project_id}-${var.environment}"
  location      = var.region
  force_destroy = var.environment != "production"
  
  uniform_bucket_level_access = true
  
  versioning {
    enabled = var.environment == "production"
  }
  
  lifecycle_rule {
    condition {
      age = 90
    }
    action {
      type          = "SetStorageClass"
      storage_class = "NEARLINE"
    }
  }
  
  lifecycle_rule {
    condition {
      age = 365
    }
    action {
      type          = "SetStorageClass"
      storage_class = "COLDLINE"
    }
  }
  
  cors {
    origin          = ["*"]
    method          = ["GET", "HEAD", "PUT", "POST", "DELETE"]
    response_header = ["*"]
    max_age_seconds = 3600
  }
}

# Secret Manager secrets
resource "google_secret_manager_secret" "jwt_secret" {
  secret_id = "agro-mas-jwt-secret"
  
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "jwt_secret_version" {
  secret      = google_secret_manager_secret.jwt_secret.id
  secret_data = var.jwt_secret
}

resource "google_secret_manager_secret" "db_password" {
  secret_id = "agro-mas-db-password"
  
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "db_password_version" {
  secret      = google_secret_manager_secret.db_password.id
  secret_data = var.db_password
}

# Service account for Cloud Run
resource "google_service_account" "agro_mas_backend" {
  account_id   = "agro-mas-backend"
  display_name = "Agro Mas Backend Service Account"
  description  = "Service account for Agro Mas backend Cloud Run service"
}

# IAM roles for the service account
resource "google_project_iam_member" "cloud_sql_client" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.agro_mas_backend.email}"
}

resource "google_project_iam_member" "storage_admin" {
  project = var.project_id
  role    = "roles/storage.admin"
  member  = "serviceAccount:${google_service_account.agro_mas_backend.email}"
}

resource "google_project_iam_member" "secret_accessor" {
  project = var.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${google_service_account.agro_mas_backend.email}"
}

# Cloud Build trigger
resource "google_cloudbuild_trigger" "agro_mas_trigger" {
  name        = "agro-mas-backend-trigger"
  description = "Trigger for Agro Mas backend deployment"
  
  github {
    owner = var.github_owner
    name  = var.github_repo
    push {
      branch = var.environment == "production" ? "^main$" : "^develop$"
    }
  }
  
  filename = "deployments/cloudbuild.yaml"
  
  substitutions = {
    _DB_HOST           = google_sql_database_instance.agro_mas_db.private_ip_address
    _DB_USER           = google_sql_user.agro_mas_user.name
    _DB_NAME           = google_sql_database.agro_mas_database.name
    _CLOUD_SQL_INSTANCE = google_sql_database_instance.agro_mas_db.connection_name
    _VPC_CONNECTOR     = google_vpc_access_connector.agro_mas_connector.name
    _SERVICE_ACCOUNT   = google_service_account.agro_mas_backend.email
    _STORAGE_BUCKET    = google_storage_bucket.agro_mas_storage.name
  }
}

# Cloud Monitoring notification channel (example: email)
resource "google_monitoring_notification_channel" "email" {
  display_name = "Agro Mas Email Notifications"
  type         = "email"
  
  labels = {
    email_address = var.notification_email
  }
}

# Cloud Monitoring alert policy for high error rate
resource "google_monitoring_alert_policy" "high_error_rate" {
  display_name = "Agro Mas - High Error Rate"
  
  conditions {
    display_name = "Cloud Run service error rate"
    
    condition_threshold {
      filter          = "resource.type=\"cloud_run_revision\" AND resource.labels.service_name=\"agro-mas-backend\""
      duration        = "300s"
      comparison      = "COMPARISON_GREATER_THAN"
      threshold_value = 0.05
      
      aggregations {
        alignment_period   = "60s"
        per_series_aligner = "ALIGN_RATE"
      }
    }
  }
  
  notification_channels = [google_monitoring_notification_channel.email.name]
  
  alert_strategy {
    auto_close = "1800s"
  }
}

# Cloud Monitoring alert policy for high latency
resource "google_monitoring_alert_policy" "high_latency" {
  display_name = "Agro Mas - High Latency"
  
  conditions {
    display_name = "Cloud Run service latency"
    
    condition_threshold {
      filter          = "resource.type=\"cloud_run_revision\" AND resource.labels.service_name=\"agro-mas-backend\""
      duration        = "300s"
      comparison      = "COMPARISON_GREATER_THAN"
      threshold_value = 2000
      
      aggregations {
        alignment_period   = "60s"
        per_series_aligner = "ALIGN_DELTA"
      }
    }
  }
  
  notification_channels = [google_monitoring_notification_channel.email.name]
  
  alert_strategy {
    auto_close = "1800s"
  }
}