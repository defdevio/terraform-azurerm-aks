resource "azurerm_kubernetes_cluster" "aks" {
  count               = var.resource_count > 0 ? 1 : 0
  name                = "${var.environment}-${var.location}-${var.name}-aks"
  location            = var.location
  resource_group_name = var.resource_group_name

  dns_prefix          = var.dns_prefix
  kubernetes_version  = var.kubernetes_version
  oidc_issuer_enabled = true
  sku_tier            = var.environment == "dev" ? "Free" : "Paid"
  tags                = var.tags

  azure_active_directory_role_based_access_control {
    managed                = true
    admin_group_object_ids = var.admin_group_object_ids
  }

  default_node_pool {
    enable_auto_scaling = var.enable_auto_scaling
    max_count           = var.enable_auto_scaling ? var.max_node_count : null
    max_pods            = var.max_pods
    min_count           = var.enable_auto_scaling ? var.min_node_count : null
    name                = var.name
    node_count          = var.node_count
    vm_size             = var.vm_size
  }

  network_profile {
    network_plugin      = "azure"
    network_plugin_mode = "Overlay"
    pod_cidr            = "192.168.0.0/16"
  }

  identity {
    type = "SystemAssigned"
  }

  lifecycle {
    ignore_changes = [
      default_node_pool[0].node_count
    ]
  }
}

resource "azurerm_kubernetes_cluster_node_pool" "node_pool" {
  for_each = { for k, v in var.additional_node_pools : k => v if var.resource_count > 0 }
  name     = each.key

  enable_auto_scaling   = true
  kubernetes_cluster_id = try(azurerm_kubernetes_cluster.aks[0].id, null)
  max_count             = try(each.value.max_node_count, null)
  max_pods              = 60
  min_count             = try(each.value.min_node_count, null)
  node_count            = each.value.node_count
  node_taints           = try(each.value.node_taints, null)
  orchestrator_version  = each.value.orchestrator_version
  tags                  = try(each.value.tags, null)
  vm_size               = each.value.vm_size

  lifecycle {
    ignore_changes = [
      node_count
    ]
  }
}