terraform {
  required_providers {
    homelab = {
      source = "registry.terraform.io/sflab-io/homelab"
      version = ">= 0.2.0"
    }
  }
}

provider "homelab" {}

# Example: Development web server
data "homelab_naming" "dev_web" {
  env = "dev"
  app = "web"
}

# Example: prod / production database
data "homelab_naming" "prod_db" {
  env = "prod"
  app = "db"
}

data "homelab_naming" "production_db" {
  env = "production"
  app = "db"
}

# Example: Staging API
data "homelab_naming" "staging_api" {
  env = "staging"
  app = "api"
}

output "dev_web_name" {
  value       = data.homelab_naming.dev_web.name
  description = "Generated name for dev web server"
}

output "prod_db_name" {
  value       = data.homelab_naming.prod_db.name
  description = "Generated name for prod database"
}

output "production_db_name" {
  value       = data.homelab_naming.production_db.name
  description = "Generated name for production database"
}

output "staging_api_name" {
  value       = data.homelab_naming.staging_api.name
  description = "Generated name for staging API"
}
