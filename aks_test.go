package tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2020-10-01/resources"
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
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
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
		"dns_prefix":          "defdevio-test",
		"environment":         "test",
		"location":            "westus",
		"name":                "cluster",
		"resource_count":      1,
		"resource_group_name": "test",
	}

	return testVars
}

// Creates a file if does not already exist on the specified path
func CreateFile(path string, content string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			bytes := []byte(content)
			err = os.WriteFile(path, bytes, 0644)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Gets the current working directory of the provider.tf file
func GetProviderPath(t *testing.T) string {
	providerFilePath := path.Join(workDir, "provider.tf")
	return providerFilePath
}

// Creates the required provider file on the system
func CreateProviderFile(providerFilePath string, t *testing.T) error {
	providerContent := `
provider "azurerm" {
	features {}
}`

	err := CreateFile(providerFilePath, providerContent)
	return err
}

// Cleans up the common files created by terraform during init, plan, and apply
func CleanUpTestFiles(t *testing.T, files []string) error {
	for _, file := range files {
		filePath := path.Join(workDir, file)
		err := os.RemoveAll(filePath)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestCreateAKSClusterWithNodePool(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "",
		Vars:         terraformVars(),
	})

	// Get the path to the provider file
	providerFile := GetProviderPath(t)

	// Defer deleting the provider file until all test functions have completed
	defer os.Remove(providerFile)

	// Create the provider file
	err := CreateProviderFile(providerFile, t)
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
	// additional_node_pools map and store them in the nodeKeys slice
	nodeKeys := make([]string, len(additionalNodePools))
	i := 0
	for key := range additionalNodePools {
		nodeKeys[i] = key
		i++
	}

	// Create a resource group client
	resourceGroupClient, err := azure.CreateResourceGroupClientE(subscriptionID)
	if err != nil {
		t.Fatal(err)
	}

	// Defer the deletion of the resource group until all test functions have finished
	defer resourceGroupClient.Delete(ctx, resourceGroup)

	// Create the resource group using the resourceGroupClient
	resp, err := resourceGroupClient.CreateOrUpdate(ctx, resourceGroup, resources.Group{
		Location: &location,
	})
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode == 201 {
		log.Printf("Created resource group '%s'", *resp.Name)
	}

	// Defer cleaning up the test files created during the test
	files := []string{"terraform.tfstate",
		"terraform.tfstate.backup",
		".terraform.lock.hcl",
		".terraform",
		".kube",
		"config",
	}

	defer CleanUpTestFiles(t, files)

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
				break
			}
		}
	}
}
