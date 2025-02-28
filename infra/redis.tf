// resource "kubernetes_namespace" "demo" {
//   metadata {
//     name = "namespace-by-tf"
//   }
// }

resource "kubernetes_service" "redisdemo" {
  metadata {
    name = "service-redis-by-tf"
  }

  spec {
    selector = {
      app = "TerraformRedisDemo"
    }
    session_affinity = "ClientIP"
    port {
      port        = 6379
      target_port = 6379
    }

    type = "NodePort"
  }
}

resource "kubernetes_deployment" "redisdemo" {
  metadata {
    name = "terraform-redis-demo"
    labels = {
      app = "TerraformRedisDemo"
    }
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "TerraformRedisDemo"
      }
    }

    template {
      metadata {
        labels = {
          app = "TerraformRedisDemo"
        }
      }

      spec {
        container {
          image = "redis:latest"
          name  = "redisdemo"

          resources {
            limits = {
              cpu    = "0.5"
              memory = "512Mi"
            }
            requests = {
              cpu    = "250m"
              memory = "50Mi"
            }
          }
        }
      }
    }
  }
}
