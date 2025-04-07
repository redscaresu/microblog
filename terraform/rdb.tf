resource "scaleway_iam_application" "blog" {
  name     = "blog"
}

resource "scaleway_iam_policy" "db_access" {
  name           = "my policy"
  description    = "gives app access to serverless database in project"
  application_id = scaleway_iam_application.blog.id
  rule {
    project_ids          = [data.scaleway_account_project.default.id]
    permission_set_names = ["ServerlessSQLDatabaseReadWrite"]
  }
}

resource "scaleway_iam_api_key" "api_key" {
  application_id = scaleway_iam_application.blog.id
}

resource "scaleway_sdb_sql_database" "blog" {
  name     = "blog"
  min_cpu  = 0
  max_cpu  = 1
}

output "database_connection_string" {
  value = format("postgres://%s:%s@%s",
    scaleway_iam_application.blog.id,
    scaleway_iam_api_key.api_key.secret_key,
    trimprefix(scaleway_sdb_sql_database.blog.endpoint, "postgres://"),
  )
  sensitive = true
}

output "database_id" {
  value     = scaleway_sdb_sql_database.blog.id
  sensitive = false
}