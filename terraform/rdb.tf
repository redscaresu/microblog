data "scaleway_account_project" "default" {
  name            = "default"
  provider = scaleway.p2
}

resource scaleway_iam_application "blog" {
  name = "blog"
  provider = scaleway.p2
}

resource scaleway_iam_policy "db_access" {
  provider = scaleway.p2
  name = "my policy"
  description = "gives app access to serverless database in project"
  application_id = scaleway_iam_application.blog.id
  rule {
    project_ids = [data.scaleway_account_project.default.id]
    permission_set_names = ["ServerlessSQLDatabaseReadWrite"]
  }
}

resource scaleway_iam_api_key "api_key" {
  provider = scaleway.p2
  application_id = scaleway_iam_application.blog.id
}

resource scaleway_sdb_sql_database "blog" {
  provider = scaleway.p2
  name = "blog"
  min_cpu = 0
  max_cpu = 1
}

output "database_connection_string" {
  value = format("postgres://%s:%s@%s",
    scaleway_iam_application.blog.id,
    scaleway_iam_api_key.api_key.secret_key,
    trimprefix(scaleway_sdb_sql_database.blog.endpoint, "postgres://"),
  )
  sensitive = true
}