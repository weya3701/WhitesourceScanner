resource "kubernetes_config_map" "nexuspropertiesdemo" {
  metadata {
    name = "nexus-properties"
  }

  data = {
    "nexus.properties" = "${file("./conf/nexus.properties")}"
  }
}
