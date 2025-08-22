package skewer

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
)

// ResourceSKUsClient is the required Azure track2 client interface used to populate skewer's data.
type ResourceSKUsClient interface {
	NewListPager(options *armcompute.ResourceSKUsClientListOptions) *runtime.Pager[armcompute.ResourceSKUsClientListResponse]
}

var _ ResourceSKUsClient = &armcompute.ResourceSKUsClient{}

// client defines the internal interface required by the skewer Cache.
type client interface {
	List(ctx context.Context, filter, includeExtendedLocations string) ([]*armcompute.ResourceSKU, error)
}
