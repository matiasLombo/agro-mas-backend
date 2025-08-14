# Terraform variables for Agro Mas Backend

variable "project_id" {
  description = "The GCP project ID"
  type        = string
}

variable "region" {
  description = "The GCP region"
  type        = string
  default     = "us-central1"
}

variable "zone" {
  description = "The GCP zone"
  type        = string
  default     = "us-central1-a"
}

variable "environment" {
  description = "Environment name (development, staging, production)"
  type        = string
  default     = "development"
  
  validation {
    condition     = contains(["development", "staging", "production"], var.environment)
    error_message = "Environment must be one of: development, staging, production."
  }
}

# Database variables
variable "db_tier" {
  description = "Cloud SQL instance tier"
  type        = string
  default     = "db-f1-micro"
  
  validation {
    condition = contains([
      "db-f1-micro", "db-g1-small", "db-n1-standard-1", "db-n1-standard-2",
      "db-n1-standard-4", "db-n1-standard-8", "db-n1-standard-16",
      "db-n1-highmem-2", "db-n1-highmem-4", "db-n1-highmem-8", "db-n1-highmem-16"
    ], var.db_tier)
    error_message = "Invalid database tier specified."
  }
}

variable "db_disk_size" {
  description = "Cloud SQL disk size in GB"
  type        = number
  default     = 20
  
  validation {
    condition     = var.db_disk_size >= 10 && var.db_disk_size <= 30720
    error_message = "Database disk size must be between 10 and 30720 GB."
  }
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}

# JWT secret
variable "jwt_secret" {
  description = "JWT secret key"
  type        = string
  sensitive   = true
}

# GitHub repository information for Cloud Build
variable "github_owner" {
  description = "GitHub repository owner"
  type        = string
}

variable "github_repo" {
  description = "GitHub repository name"
  type        = string
  default     = "agro-mas-backend"
}

# Notification settings
variable "notification_email" {
  description = "Email address for monitoring notifications"
  type        = string
}

# Cloud Run settings
variable "cloud_run_min_instances" {
  description = "Minimum number of Cloud Run instances"
  type        = number
  default     = 0
  
  validation {
    condition     = var.cloud_run_min_instances >= 0 && var.cloud_run_min_instances <= 10
    error_message = "Minimum instances must be between 0 and 10."
  }
}

variable "cloud_run_max_instances" {
  description = "Maximum number of Cloud Run instances"
  type        = number
  default     = 100
  
  validation {
    condition     = var.cloud_run_max_instances >= 1 && var.cloud_run_max_instances <= 1000
    error_message = "Maximum instances must be between 1 and 1000."
  }
}

variable "cloud_run_memory" {
  description = "Memory allocation for Cloud Run instances"
  type        = string
  default     = "2Gi"
  
  validation {
    condition = contains([
      "128Mi", "256Mi", "512Mi", "1Gi", "2Gi", "4Gi", "8Gi", "16Gi", "32Gi"
    ], var.cloud_run_memory)
    error_message = "Invalid memory allocation. Must be one of: 128Mi, 256Mi, 512Mi, 1Gi, 2Gi, 4Gi, 8Gi, 16Gi, 32Gi."
  }
}

variable "cloud_run_cpu" {
  description = "CPU allocation for Cloud Run instances"
  type        = string
  default     = "2"
  
  validation {
    condition = contains([
      "0.08", "0.17", "0.33", "0.5", "1", "2", "4", "6", "8"
    ], var.cloud_run_cpu)
    error_message = "Invalid CPU allocation. Must be one of: 0.08, 0.17, 0.33, 0.5, 1, 2, 4, 6, 8."
  }
}

# Domain settings
variable "domain_name" {
  description = "Custom domain name for the API (optional)"
  type        = string
  default     = ""
}

# Enable/disable features
variable "enable_monitoring" {
  description = "Enable Cloud Monitoring and alerting"
  type        = bool
  default     = true
}

variable "enable_logging" {
  description = "Enable structured logging"
  type        = bool
  default     = true
}

variable "enable_backup" {
  description = "Enable database backups"
  type        = bool
  default     = true
}

# Cost optimization settings
variable "storage_lifecycle_enabled" {
  description = "Enable storage lifecycle management"
  type        = bool
  default     = true
}

variable "preemptible_instances" {
  description = "Use preemptible instances for non-production environments"
  type        = bool
  default     = false
}

# Security settings
variable "require_ssl" {
  description = "Require SSL for database connections"
  type        = bool
  default     = true
}

variable "enable_private_ip" {
  description = "Use private IP for Cloud SQL"
  type        = bool
  default     = true
}

# Scaling settings based on environment
locals {
  environment_config = {
    development = {
      db_tier           = "db-f1-micro"
      db_disk_size      = 10
      min_instances     = 0
      max_instances     = 10
      memory           = "512Mi"
      cpu              = "1"
      backup_enabled   = false
    }
    staging = {
      db_tier           = "db-g1-small"
      db_disk_size      = 20
      min_instances     = 1
      max_instances     = 20
      memory           = "1Gi"
      cpu              = "1"
      backup_enabled   = true
    }
    production = {
      db_tier           = "db-n1-standard-2"
      db_disk_size      = 100
      min_instances     = 2
      max_instances     = 100
      memory           = "2Gi"
      cpu              = "2"
      backup_enabled   = true
    }
  }
  
  config = local.environment_config[var.environment]
}