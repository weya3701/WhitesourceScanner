resource "helm_release" "kube_state_metrics" {
  name = "kube-state-metrics"

  repository = "https://kubernetes.github.io/kube-state-metrics"
  chart      = "kube-state-metrics"

  set {
    name  = "service.type"
    value = "NodePort"
  }
}
