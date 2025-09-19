package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"

	"github.com/Azure/skewer/v2"
)

func main() {
	// az login
	// export AZURE_SUBSCRIPTION_ID="subscription-id"
	sub := os.Getenv("AZURE_SUBSCRIPTION_ID")
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		fmt.Printf("failed to get credential: %s", err)
		os.Exit(1)
	}

	client, err := armcompute.NewResourceSKUsClient(sub, cred, nil)
	if err != nil {
		fmt.Printf("failed to get client: %s", err)
		os.Exit(1)
	}

	cache, err := skewer.NewCache(context.Background(), skewer.WithLocation("eastus"), skewer.WithResourceSKUsClient(client))
	if err != nil {
		fmt.Printf("failed to instantiate sku cache: %s", err)
		os.Exit(1)
	}

	for _, sku := range cache.List(context.Background()) {
		fmt.Printf("sku: %s\n", sku.GetName())
	}

	err = checkSKU(cache)
	if err != nil {
		fmt.Printf("failed to check sku: %s", err)
		os.Exit(1)
	}
}

func checkSKU(cache *skewer.Cache) error {
	sku, err := cache.Get(context.Background(), "standard_d4s_v3", skewer.VirtualMachines, "eastus")
	if err != nil {
		return fmt.Errorf("failed to find virtual machine sku standard_d4s_v3: %s", err)
	}

	// Check for capabilities
	if sku.IsEphemeralOSDiskSupported() {
		fmt.Printf("SKU %s supports ephemeral OS disk!\n", sku.GetName())
	}

	cpu, err := sku.VCPU()
	if err != nil {
		return fmt.Errorf("failed to parse cpu from sku: %s", err)
	}

	memory, err := sku.Memory()
	if err != nil {
		return fmt.Errorf("failed to parse memory from sku: %s", err)
	}

	fmt.Printf("vm sku %s has %d vCPU cores and %.2fGi of memory\n", sku.GetName(), cpu, memory)

	return nil
}
