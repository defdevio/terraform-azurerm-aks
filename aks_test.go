package tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2020-10-01/resources"
	"github.com/gruntwork-io/terratest/modules/azure"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

var (
	ctx            = context.Background()
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
)

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
func GetProviderPath(t *testing.T) (string, error) {
	workDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	providerFilePath := path.Join(workDir, "provider.tf")
	return providerFilePath, err
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
func CleanUpTerraformFiles(t *testing.T) error {
	workDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	tfFiles := []string{"terraform.tfstate", "terraform.tfstate.backup", "terraform.lock.hcl", ".terraform"}
	for _, file := range tfFiles {
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
		Vars: map[string]any{
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
		},
	})

	// Get the path to the provider file
	providerFile, err := GetProviderPath(t)
	if err != nil {
		t.Fatal(err)
	}

	// Defer deleting the provider file until all test functions have completed
	defer os.Remove(providerFile)

	// Create the provider file
	err = CreateProviderFile(providerFile, t)
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

	// Defer destroying the terraform resources until the rest of the test functions finish
	defer terraform.Destroy(t, terraformOptions)

	// Defer cleaning up the terraform files created during init, plan, and apply
	defer CleanUpTerraformFiles(t)

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

	adminKubeConfig := terraform.Output(t, terraformOptions, "admin_kube_config")
	t.Logf("The admin_kube_config is: %s", adminKubeConfig)
}
