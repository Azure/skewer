package main

import (
	"context"
	"fmt"
	"os"
	"text/template"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute" //nolint:staticcheck
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/skewer/testdata"
)

func getSKUs(subscriptionID, region string) (map[string]testdata.SKUInfo, error) {
	authorizer, err := auth.NewAuthorizerFromCLI()
	if err != nil {
		return nil, err
	}

	// Create a new compute client
	client := compute.NewResourceSkusClient(subscriptionID)
	client.Authorizer = authorizer

	// List SKUs for the specified region
	skuList, err := client.List(context.Background(), region, "")
	if err != nil {
		return nil, err
	}

	skus := map[string]testdata.SKUInfo{}
	for _, sku := range skuList.Values() {
		if *sku.ResourceType == "virtualMachines" {
			if _, ok := skus[*sku.Name]; !ok {
				skuInfo := testdata.SKUInfo{
					Size: *sku.Size,
				}
				skus[*sku.Name] = skuInfo
			}
		}
	}

	return skus, nil
}

const templateCode = `package testdata

type SKUInfo struct {
	Size string
}

var SKUData = map[string]SKUInfo{
{{- range $key, $value := .}}
	"{{ $key }}": {
		Size:   "{{ $value.Size }}",
	},
{{- end }}
}
`

func generateAndSaveFile(skus map[string]testdata.SKUInfo) error {
	file, err := os.Create("../testdata/generated_vmsize_testdata.go")
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl, err := template.New("skus_template").Parse(templateCode)
	if err != nil {
		return err
	}

	err = tmpl.Execute(file, skus)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	// Get the subscription ID from the environment variable or use a default value
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")

	// Get the region from the environment variable or use a default value
	region := os.Getenv("AZURE_REGION")
	if region == "" {
		region = "eastus" // Default region if not provided in the environment variable
	}

	skus, err := getSKUs(subscriptionID, region)
	if err != nil {
		fmt.Println("Error fetching SKUs:", err)
		return
	}

	err = generateAndSaveFile(skus)
	if err != nil {
		fmt.Println("Error generating and saving file:", err)
		return
	}

	fmt.Println("Generated and saved skudata/skus_generated.go successfully!")
}
