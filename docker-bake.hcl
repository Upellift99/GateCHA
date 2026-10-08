variable "TAG" {
  default = "latest"
}

variable "REGISTRY" {
  default = "ghcr.io/upellift99"
}

variable "VERSION" {
  default = "dev"
}

# Placeholders that docker/metadata-action's bake files override in CI with the
# real tags and labels. The defaults here only apply to local `docker buildx bake`.
target "docker-metadata-action" {
  tags = ["${REGISTRY}/gatecha:${TAG}"]
}

target "docker-metadata-action-mysql" {
  tags = ["${REGISTRY}/gatecha:${TAG}-mysql"]
}

group "default" {
  targets = ["app", "mysql"]
}

target "app" {
  inherits   = ["docker-metadata-action"]
  dockerfile = "Dockerfile"
  args = {
    VERSION = "${VERSION}"
  }
  platforms = ["linux/amd64", "linux/arm64"]
}

target "mysql" {
  inherits   = ["docker-metadata-action-mysql"]
  dockerfile = "Dockerfile"
  args = {
    VERSION    = "${VERSION}"
    BUILD_TAGS = "mysql"
  }
  platforms = ["linux/amd64", "linux/arm64"]
}
