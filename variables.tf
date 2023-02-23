variable "additional_node_pools" {
  type = map(any)
  default = {
    example = {
      max_node_count       = 4
      min_node_count       = 2
      node_count           = 2
      orchestrator_version = "1.25.0"
      vm_size              = "Standard_B1ms"
    }
  }
  description = "A map that describes the configuration for additional node pools."
}

variable "admin_group_object_ids" {
  type        = list(string)
  default     = []
  description = "A list of Azure AD `user`/`group`/`service principals` that will be provided `ClusterAdmin` rights."
}

variable "resource_count" {
  type        = number
  default     = 0
  description = "The number of AKS clusters to provision."
}

variable "dns_prefix" {
  type        = string
  description = "The DNS prefix for the AKS cluster."
}

variable "enable_auto_scaling" {
  type        = bool
  default     = true
  description = "A flag to enable node auto scaling."
}

variable "environment" {
  type        = string
  description = "The name of the deployment environment for the resource."
}

variable "location" {
  type        = string
  description = "The Azure region where the resource will be provisioned."
}

variable "kubernetes_version" {
  type        = string
  default     = "1.25.0"
  description = "The version of Kubernetes."
}

variable "name" {
  type        = string
  description = "The name to provide to the resource."
}

variable "node_count" {
  type        = number
  default     = 2
  description = "The number of nodes to initially provision."
}

variable "max_node_count" {
  type        = number
  default     = 4
  description = "The maximum number of nodes availalbe in the auto scaler pool."
}

variable "max_pods" {
  type        = number
  default     = 30
  description = "The maximum number of pods available to deploy on a node."
}

variable "min_node_count" {
  type        = number
  default     = 1
  description = "The minimum number of nodes available in the auto scaler pool."
}

variable "resource_group_name" {
  type        = string
  description = "The Azure resource group where the resources will be provisioned."
}

variable "tags" {
  type        = map(string)
  default     = {}
  description = "A map of tags to apply to the resources."
}

variable "vm_size" {
  type        = string
  default     = "Standard_B1ms"
  description = "The VM sku to use for the `default` node pool"
}