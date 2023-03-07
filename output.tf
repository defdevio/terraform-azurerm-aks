output "admin_kube_config" {
  value       = try(azurerm_kubernetes_cluster.aks[0].kube_admin_config_raw, null)
  sensitive   = true
  description = "The admin kubeconfig for the AKS cluster."
}