# skewer [![GoDoc](https://godoc.org/github.com/Azure/skewer?status.svg)](https://godoc.org/github.com/Azure/skewer) [![codecov](https://codecov.io/gh/azure/skewer/branch/master/graph/badge.svg)](https://codecov.io/gh/azure/skewer)

A package to simplify working with Azure's Resource SKU APIs by wrapping
the existing Azure SDK for Go.

## Usage

This package requires an existing, authorized Azure client. Here is a
complete example using the simplest methods.

```go
package main

import (
    "context"
    "fmt"

    "github.com/Azure/go-autorest/autorest/azure/auth"
    "github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"

    "github.com/Azure/skewer"
)

func main() {
    authorizer, err := auth.NewAuthorizerFromEnvironment()
    if err != nil {
        fmt.Printf("failed to get authorizer: %s", err)
        os.Exit(1)
    }
	client := compute.NewResourceSkusClient(sub)
    client.Authorizer = authorizer
    // Now we can use the client...
    resourceSkuIterator, err := client.ListComplete(context.Background(), "eastus")
    if err != nil {
        fmt.Printf("failed to list skus: %s", err)
            os.Exit(1)
        }
    // or instantiate a cache for this package!
    cache, err := skewer.NewCache(context.Background(), skewer.WithLocation("eastus"), skewer.WithResourceClient(client))
    if err != nil {
        fmt.Printf("failed to instantiate sku cache: %s", err)
        os.Exit(1)
    }
}
```

Once we have a cache, we can query against its contents:
```go
sku, found := cache.Get(context.Background, "standard_d4s_v3", skewer.VirtualMachines)
if !found {
    return fmt.Errorf("expected to find virtual machine sku standard_d4s_v3")
}

// Check for capabilities
if sku.IsEphemeralOSDiskSupported() {
    fmt.Println("SKU %s supports ephemeral OS disk!", sku.GetName())
}

cpu, err := sku.VCPU()
if err != nil {
    return fmt.Errorf("failed to parse cpu from sku: %s", err)
}

memory, err := sku.Memory()
if err != nil {
    return fmt.Errorf("failed to parse memory from sku: %s", err)
}

fmt.Printf("vm sku %s has %d vCPU cores and %.2fGi of memory", sku.GetName(), cpu, memory)
```

# Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
