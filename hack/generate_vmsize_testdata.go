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

func getSKUs(subscriptionID, region string) (map[string]testdata.SKUInfo, []testdata.SKU, error) {
	authorizer, err := auth.NewAuthorizerFromCLI()
	if err != nil {
		return nil, nil, err
	}

	// Create a new compute client
	client := compute.NewResourceSkusClient(subscriptionID)
	client.Authorizer = authorizer

	// List SKUs for the specified region
	skuList, err := client.List(context.Background(), region, "")
	if err != nil {
		return nil, nil, err
	}

	skuInfoMap := map[string]testdata.SKUInfo{}
	skuSlice := []testdata.SKU{}
	for _, sku := range skuList.Values() {
		if *sku.ResourceType == "virtualMachines" {
			if _, ok := skuInfoMap[*sku.Name]; !ok {
				skuSlice = append(skuSlice, testdata.SKU{Name: sku.Name})
				skuInfo := testdata.SKUInfo{
					Size: *sku.Size,
				}
				skuInfoMap[*sku.Name] = skuInfo
			}
		}
	}

	return skuInfoMap, skuSlice, nil
}

const vmSizeTestDataTemplate = `package testdata

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

const vmSKUsTestDataTemplate = `package testdata

import (
	"k8s.io/utils/ptr"
)

type SKU struct {
	Name *string
}

var SKUs = []SKU{
{{- range $i, $data := .}}
	{
		Name:     ptr.To("{{ $data.Name }}"),
	},
{{- end }}
}
`

func generateAndSaveFiles(skuInfoMap map[string]testdata.SKUInfo, skuSlice []testdata.SKU) error {
	vmSizeGeneratedFilename := "testdata/generated_vmsize_testdata.go"
	file, err := os.Create(fmt.Sprintf("../%s", vmSizeGeneratedFilename))
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl, err := template.New("vmSize_test_data_template").Parse(vmSizeTestDataTemplate)
	if err != nil {
		return err
	}

	err = tmpl.Execute(file, skuInfoMap)
	if err != nil {
		return err
	}
	fmt.Printf("Generated and saved %s successfully!\n", vmSizeGeneratedFilename)

	vmSKUsGeneratedFilename := "testdata/generated_vmskus_testdata.go"
	file, err = os.Create(fmt.Sprintf("../%s", vmSKUsGeneratedFilename))
	if err != nil {
		return err
	}
	defer file.Close()
	tmpl, err = template.New("vmSKUs_test_data_template").Parse(vmSKUsTestDataTemplate)
	if err != nil {
		return err
	}

	err = tmpl.Execute(file, skuSlice)
	if err != nil {
		return err
	}
	fmt.Printf("Generated and saved %s successfully!\n", vmSKUsGeneratedFilename)

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

	skuInfoMap, skuSlice, err := getSKUs(subscriptionID, region)
	if err != nil {
		fmt.Println("Error fetching SKUs:", err)
		return
	}

	err = generateAndSaveFiles(skuInfoMap, skuSlice)
	if err != nil {
		fmt.Println("Error generating and saving file:", err)
		return
	}
}
