module "telemetry" {
  source              = "github.com/defdevio/terraform-azurerm-telemetry?ref=latest"
  resource_count      = var.resource_count > 0 ? 1 : 0
  resource_group_name = var.resource_group_name

  environment = var.environment
  location    = var.location
  tags        = var.tags

  create_law          = var.create_telemetry_law
  law_resource_id     = var.law_resource_id
  log_categories      = try(data.azurerm_monitor_diagnostic_categories.categories[0].log_category_types, [])
  metric_categories   = try(data.azurerm_monitor_diagnostic_categories.categories[0].metrics, [])
  target_resource_ids = azurerm_kubernetes_cluster.aks[*].id
}

data "azurerm_monitor_diagnostic_categories" "categories" {
  count       = var.resource_count > 0 ? 1 : 0
  resource_id = azurerm_kubernetes_cluster.aks[count.index].id
}