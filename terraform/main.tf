resource "scaleway_container_namespace" "main" {
  name        = "blog"
  description = "blog"
  project_id  = var.project_id
}

resource "random_password" "auth_password" {
  length           = 16
  special          = true
  min_numeric      = 1
  min_upper        = 1
  min_lower        = 1
  min_special      = 1
  override_special = "_-"
}

resource "scaleway_container" "main" {
  name           = "blog"
  description    = "my blog"
  namespace_id   = scaleway_container_namespace.main.id
  registry_image = "${scaleway_container_namespace.main.registry_endpoint}/microblog:latest"
  port           = 8080
  cpu_limit      = 70
  memory_limit   = 128
  min_scale      = 1
  max_scale      = 1
  timeout        = 600
  privacy        = "public"
  protocol       = "http1"
  deploy         = true
  http_option    = "redirected"

  scaling_option {
    concurrent_requests_threshold = 10
  }

  environment_variables = {
    "AUTH_USERNAME" = "admin",
  }

  secret_environment_variables = {
    "DB_PASSWORD"   = scaleway_iam_api_key.api_key.secret_key
    "DB_USER"       = scaleway_iam_application.blog.id,
    "DB_HOST"       = trimsuffix(trimprefix(regex(":\\/\\/.*:", scaleway_sdb_sql_database.blog.endpoint), "://"), ":")
    "DB_NAME"       = scaleway_sdb_sql_database.blog.name,
    "DB_PORT"       = trimprefix(regex(":[0-9]{1,5}", scaleway_sdb_sql_database.blog.endpoint), ":"),
    "DB_ID"         = scaleway_sdb_sql_database.blog.id
    "AUTH_PASSWORD" = random_password.auth_password.result
  }
}

resource "scaleway_container_domain" "main" {
  container_id = scaleway_container.main.id
  hostname     = "ashouri.xyz"
}

output "auth_password" {
  value       = random_password.auth_password.result
  sensitive   = true
  description = "The generated random password for the exporter"
}
