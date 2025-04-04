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
    port = 9997
    cpu_limit = 70
    memory_limit = 128
    min_scale = 1
    max_scale = 1
    timeout = 600
    privacy = "public"
    protocol = "http1"
    deploy = true
    http_option = "redirected"

    environment_variables = {
        "foo" = "var"
    }
    secret_environment_variables = {
      "key" = "secret"
    }
}