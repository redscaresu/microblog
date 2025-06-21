variable "organization_id" {
  description = "The Scaleway organization ID"
  type        = string
}

variable "project_id" {
  description = "The Scaleway project ID"
  type        = string
}

variable "container_image_tag" {
  description = "The container image tag to deploy"
  type        = string
  default     = "latest"
}