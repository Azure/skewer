package skewer

import (
	"context"
	"fmt"
)

func Match(ctx context.Context, cache *Cache, sku *SKU, location string) *SKU {
	sizes := cache.List(ctx, ResourceTypeFilter(sku.GetResourceType()), LocationFilter(normalizeLocation(location)))

	capabilities := map[string]string{}
	for _, capability := range *sku.Capabilities {
		if capability.Name != nil {
			if capability.Value != nil {
				capabilities[*capability.Name] = *capability.Value
			} else {
				capabilities[*capability.Name] = ""
			}
		}
	}

	for i := range sizes {
		candidate := &sizes[i]
		if candidate.GetName() == sku.GetName() {
			continue
		}
		if allCapabilitiesMatch(candidate, capabilities) {
			return candidate
		}
	}

	return nil
}

func allCapabilitiesMatch(sku *SKU, capabilities map[string]string) bool {
	matched := 0
	desired := len(capabilities)
	for _, capability := range *sku.Capabilities {
		if capability.Name != nil {
			// TODO(ace): this is not actually accurate, really for each
			// capability, we should decide whether you need subset, exact match,
			// or numerically greater/less than.
			if capabilitiesToIgnore[*capability.Name] {
				continue
			}
			if capability.Value != nil {
				// TODO(ace): this is far too strict and results in basically zero matches.
				if capabilities[*capability.Name] != *capability.Value {
					fmt.Printf("failed on capability %s=%s\n", *capability.Name, *capability.Value)
					return false
				}
				matched++
			} else {
				val, ok := capabilities[*capability.Name]
				if !ok || val != "" {
					fmt.Printf("failed on capability %s with no value\n", *capability.Name)
					return false
				}
				matched++
			}
		}
	}

	if matched != desired {
		fmt.Printf("failed to find all desired capabilities want %d got %d\n", desired, matched)
	}
	return matched == desired
}

var capabilitiesToIgnore = map[string]bool{
	MaxResourceVolumeMB: true,
}
