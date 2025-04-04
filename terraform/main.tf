resource scaleway_container_namespace main {
    name = "blog"
    description = "blog"
    provider = scaleway.p2
}

# resource scaleway_container main {
#     provider = scaleway.p2
#     name = "my-container-02"
#     description = "environment variables test"
#     namespace_id = scaleway_container_namespace.main.id
#     registry_image = "${scaleway_container_namespace.main.registry_endpoint}/alpine:test"
#     port = 9997
#     cpu_limit = 140
#     memory_limit = 256
#     min_scale = 1
#     max_scale = 1
#     timeout = 600
#     privacy = "private"
#     protocol = "http1"
#     deploy = true

#     environment_variables = {
#         "foo" = "var"
#     }
#     secret_environment_variables = {
#       "key" = "secret"
#     }
# }