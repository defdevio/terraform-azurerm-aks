<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 0.12.26 |
| <a name="requirement_azurerm"></a> [azurerm](#requirement\_azurerm) | ~> 3.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_azurerm"></a> [azurerm](#provider\_azurerm) | ~> 3.0 |

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_telemetry"></a> [telemetry](#module\_telemetry) | github.com/defdevio/terraform-azurerm-telemetry | latest |

## Resources

| Name | Type |
|------|------|
| [azurerm_kubernetes_cluster.aks](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/kubernetes_cluster) | resource |
| [azurerm_kubernetes_cluster_node_pool.node_pool](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/kubernetes_cluster_node_pool) | resource |
| [azurerm_monitor_diagnostic_categories.categories](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/data-sources/monitor_diagnostic_categories) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_additional_node_pools"></a> [additional\_node\_pools](#input\_additional\_node\_pools) | A map that describes the configuration for additional node pools. | `map(any)` | <pre>{<br>  "example": {<br>    "max_node_count": 4,<br>    "min_node_count": 2,<br>    "node_count": 1,<br>    "orchestrator_version": "1.25.4",<br>    "vm_size": "Standard_B2ms"<br>  }<br>}</pre> | no |
| <a name="input_admin_group_object_ids"></a> [admin\_group\_object\_ids](#input\_admin\_group\_object\_ids) | A list of Azure AD `user`/`group`/`service principals` that will be provided `ClusterAdmin` rights. | `list(string)` | `[]` | no |
| <a name="input_create_telemetry_law"></a> [create\_telemetry\_law](#input\_create\_telemetry\_law) | When testing this module we create a temporary `Log Analytics Workspace` to verify logs and metrics are working. This variable acts as a flag and should be kept `false` for production usage. | `bool` | `false` | no |
| <a name="input_dns_prefix"></a> [dns\_prefix](#input\_dns\_prefix) | The DNS prefix for the AKS cluster. | `string` | n/a | yes |
| <a name="input_enable_auto_scaling"></a> [enable\_auto\_scaling](#input\_enable\_auto\_scaling) | A flag to enable node auto scaling. | `bool` | `true` | no |
| <a name="input_environment"></a> [environment](#input\_environment) | The name of the deployment environment for the resource. | `string` | n/a | yes |
| <a name="input_kubernetes_version"></a> [kubernetes\_version](#input\_kubernetes\_version) | The version of Kubernetes. | `string` | `"1.25.4"` | no |
| <a name="input_law_resource_id"></a> [law\_resource\_id](#input\_law\_resource\_id) | The resource id of the `Log Analytics Workspace` where logs and metrics should be sent for long term retention. | `string` | `""` | no |
| <a name="input_location"></a> [location](#input\_location) | The Azure region where the resource will be provisioned. | `string` | n/a | yes |
| <a name="input_max_node_count"></a> [max\_node\_count](#input\_max\_node\_count) | The maximum number of nodes availalbe in the auto scaler pool. | `number` | `4` | no |
| <a name="input_max_pods"></a> [max\_pods](#input\_max\_pods) | The maximum number of pods available to deploy on a node. | `number` | `30` | no |
| <a name="input_min_node_count"></a> [min\_node\_count](#input\_min\_node\_count) | The minimum number of nodes available in the auto scaler pool. | `number` | `1` | no |
| <a name="input_name"></a> [name](#input\_name) | The name to provide to the resource. | `string` | n/a | yes |
| <a name="input_node_count"></a> [node\_count](#input\_node\_count) | The number of nodes to initially provision. | `number` | `2` | no |
| <a name="input_resource_count"></a> [resource\_count](#input\_resource\_count) | The number of AKS clusters to provision. | `number` | `0` | no |
| <a name="input_resource_group_name"></a> [resource\_group\_name](#input\_resource\_group\_name) | The Azure resource group where the resources will be provisioned. | `string` | n/a | yes |
| <a name="input_tags"></a> [tags](#input\_tags) | A map of tags to apply to the resources. | `map(string)` | `{}` | no |
| <a name="input_vm_size"></a> [vm\_size](#input\_vm\_size) | The VM sku to use for the `default` node pool | `string` | `"Standard_B2ms"` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_admin_kube_config"></a> [admin\_kube\_config](#output\_admin\_kube\_config) | The admin kubeconfig for the AKS cluster. |
<!-- END_TF_DOCS -->