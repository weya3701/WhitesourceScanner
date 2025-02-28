// resource "kubernetes_namespace" "demo" {
//   metadata {
//     name = "namespace-by-tf"
//   }
// }

resource "kubernetes_service" "nexusdemo" {
  metadata {
    name = "service-nexus-by-tf"
  }

  spec {
    selector = {
      app = "TerraformNexusDemo"
    }
    session_affinity = "ClientIP"
    port {
      name        = "8081"
      port        = 8081
      target_port = 8081
    }
    port {
      name        = "8082"
      port        = 8082
      target_port = 8082
    }
    type = "NodePort"
  }
}

resource "kubernetes_deployment" "nexusdemo" {
  metadata {
    name = "terraform-nexus-demo"
    labels = {
      app = "TerraformNexusDemo"
    }
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "TerraformNexusDemo"
      }
    }

    template {
      metadata {
        labels = {
          app = "TerraformNexusDemo"
        }
      }

      spec {
        container {
          image = "sonatype/nexus3:latest"
          name  = "nexusdemo"
          # volume_mount {
          #   name       = "admin-pass"
          #   mount_path = "/nexus-data/"
          # }

          resources {
            limits = {
              cpu    = "4"
              memory = "4096Mi"
            }

            requests = {
              cpu    = "3.5"
              memory = "3048Mi"
            }
          }
        }
        # volume {
        #   name = "admin-pass"
        #   config_map {
        #     name = "nexus-admin-pass"
        #   }
        # }
      }
    }
  }
}
