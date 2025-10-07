package main

import (
	"context"
	"fmt"
	"os"
	"text/template"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
	"github.com/Azure/skewer/v2/testdata"
)

func getSKUs(subscriptionID, region string) (map[string]testdata.SKUInfo, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	client, err := armcompute.NewResourceSKUsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	filter := fmt.Sprintf("location eq '%s'", region)
	pager := client.NewListPager(&armcompute.ResourceSKUsClientListOptions{Filter: &filter})

	skus := map[string]testdata.SKUInfo{}
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, v := range page.Value {
			if v.ResourceType != nil && *v.ResourceType == "virtualMachines" {
				if _, ok := skus[*v.Name]; !ok {
					skuInfo := testdata.SKUInfo{
						Size: *v.Size,
					}
					skus[*v.Name] = skuInfo
				}
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
