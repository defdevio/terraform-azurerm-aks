package tests

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/defdevio/terratest-helpers/pkg/helpers"
	"github.com/gruntwork-io/terratest/modules/azure"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	ctx            = context.Background()
	kubeConfigPath = path.Join(workDir, ".kube", "config")
	subscriptionID = os.Getenv("ARM_SUBSCRIPTION_ID")
	workDir, _     = os.Getwd()
)

// The variables to pass in to the Terraform runs
func terraformVars() map[string]any {
	testVars := map[string]any{
		"additional_node_pools": map[string]any{
			"pool": map[string]any{
				"max_node_count":       2,
				"min_node_count":       1,
				"node_count":           1,
				"orchestrator_version": "1.25.4",
				"vm_size":              "Standard_B2ms",
			},
		},
		"dns_prefix":           "defdevio-test",
		"environment":          "test",
		"location":             "westus",
		"name":                 "cluster",
		"resource_count":       1,
		"resource_group_name":  "test",
		"create_telemetry_law": true,
	}

	return testVars
}

func TestCreateAKSClusterWithNodePool(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "",
		Vars:         terraformVars(),
	})

	testFiles := []string{
		"provider.tf",
		"terraform.tfstate",
		"terraform.tfstate.backup",
		".terraform.lock.hcl",
		".terraform",
		".kube",
		"config",
	}

	// Defer cleaning up the test files created during the test
	defer helpers.CleanUpTestFiles(t, testFiles, workDir)

	// Create the provider file
	err := helpers.CreateAzureProviderFile(path.Join(workDir, "provider.tf"), t)
	if err != nil {
		t.Fatal(err)
	}

	// Use type assertions to ensure the interface values are the expected type for the given
	// terraform variable value
	clusterName, ok := terraformOptions.Vars["name"].(string)
	if !ok {
		t.Fatal("A value type of 'string' was expected for 'clusterName'")
	}

	environment, ok := terraformOptions.Vars["environment"].(string)
	if !ok {
		t.Fatal("A value type of 'string' was expected for 'environment'")
	}

	location, ok := terraformOptions.Vars["location"].(string)
	if !ok {
		t.Fatal("A value type of 'string' was expected for 'location'")
	}

	resourceGroup, ok := terraformOptions.Vars["resource_group_name"].(string)
	if !ok {
		t.Fatal("A value type of 'string' was expected for 'resourceGroup'")
	}

	additionalNodePools, ok := terraformOptions.Vars["additional_node_pools"].(map[string]any)
	if !ok {
		t.Fatal("A value type of 'map[string]any' was expected for 'additionalNodePools")
	}

	// This will make a new slice array that contains the keys of the
	// additional_node_pools map
	nodeKeys := helpers.GetMapKeys(t, additionalNodePools)

	// Defer the deletion of the resource group until all test functions have finished
	defer helpers.DeleteAzureResourceGroup(t, subscriptionID, resourceGroup)

	// Create the resource group
	err = helpers.CreateAzureResourceGroup(t, subscriptionID, resourceGroup, &location)
	if err != nil {
		t.Fatal(err)
	}

	// Defer destroying the terraform resources until the rest of the test functions finish
	defer terraform.Destroy(t, terraformOptions)

	// Init and apply the terraform module
	terraform.InitAndApply(t, terraformOptions)

	// Format the cluster name to be the expected cluster name as created by the terraform module
	formattedClusterName := fmt.Sprintf("%s-%s-%s-aks", environment, location, clusterName)

	// Get the managed cluster resource the test created
	cluster, err := azure.GetManagedClusterE(t, resourceGroup, formattedClusterName, subscriptionID)
	if err != nil {
		t.Fatal(err)
	}

	// Assert that the cluster returns a succeeded provisioning state
	assert.Equal(t, "Succeeded", *cluster.ProvisioningState)

	// Assert that the deployed cluster resource has the same name as the desired resource
	assert.Equal(t, formattedClusterName, *cluster.Name)

	// Write the kubeconfig to file
	adminKubeConfig := terraform.Output(t, terraformOptions, "admin_kube_config")
	err = os.Mkdir(path.Join(workDir, ".kube"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(kubeConfigPath, []byte(adminKubeConfig), 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Create the client
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Get a list of all nodes
	for _, nodePool := range nodeKeys {
		nodes, err := clientSet.CoreV1().Nodes().List(ctx, v1.ListOptions{})
		if err != nil {
			t.Fatal(err)
		}

		// Loop through the nodes and look for the nodes that contain the node pool name
		for _, node := range nodes.Items {
			if strings.Contains(node.Name, nodePool) {
				assert.Equal(t, nodePool, node.ObjectMeta.Labels["kubernetes.azure.com/agentpool"])
			}
		}
	}
}
