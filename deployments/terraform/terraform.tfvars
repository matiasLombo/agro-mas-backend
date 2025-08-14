# Terraform variables for Agro Mas production deployment
project_id = "rising-city-422302-p9"
region     = "us-central1"
zone       = "us-central1-a"
environment = "production"

# Database configuration
db_tier      = "db-f1-micro"  # Cost-optimized for low traffic
db_disk_size = 20

# Secrets (will be set as environment variables during deployment)
db_password = "AgroMas2025SecureDB!"
jwt_secret  = "AgroMas2025JWT-Secret-Key-SuperSecure-128bit"

# GitHub repository (if using CI/CD)
github_owner = "matiasmartinez"  # Replace with actual GitHub username
github_repo  = "agro-mas"

# Notification settings
notification_email = "alerts@agro-mas.com"  # Replace with actual email

# Cloud Run settings (cost-optimized)
cloud_run_min_instances = 0
cloud_run_max_instances = 5
cloud_run_memory       = "512Mi"
cloud_run_cpu          = "1"

# Feature toggles
enable_monitoring = true
enable_logging    = true
enable_backup     = true
require_ssl       = true
enable_private_ip = false  # Simplified for initial deployment