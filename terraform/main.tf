resource scaleway_container_namespace main {
    name = "blog"
    description = "blog"
    provider = scaleway.p2
}

resource scaleway_container main {
    provider = scaleway.p2
    name = "blog"
    description = "my blog"
    namespace_id = scaleway_container_namespace.main.id
    registry_image = "${scaleway_container_namespace.main.registry_endpoint}/microblog:latest"
    port = 8080
    cpu_limit = 70
    memory_limit = 128
    min_scale = 1
    max_scale = 1
    timeout = 600
    privacy = "public"
    protocol = "http1"
    deploy = true
    http_option = "redirected"

  secret_environment_variables = {
    "DB_PASSWORD" = scaleway_iam_api_key.api_key.secret_key
    "DB_USER" = scaleway_iam_application.blog.id,
    "DB_HOST"     = trimsuffix(trimprefix(regex(":\\/\\/.*:", scaleway_sdb_sql_database.blog.endpoint), "://"), ":")
    "DB_NAME"     = scaleway_sdb_sql_database.blog.name,
    "DB_PORT"     = trimprefix(regex(":[0-9]{1,5}", scaleway_sdb_sql_database.blog.endpoint), ":"),
  }
}