package skewer

import (
	"context"
	"encoding/json"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
)

// dataWrapper is a convenience wrapper for deserializing json testdata
type dataWrapper struct {
	Value []*armcompute.ResourceSKU `json:"value,omitempty"`
}

// newDataWrapper takes a path to a list of compute skus and parses them
// to a dataWrapper for use in fake clients
func newDataWrapper(path string) (*dataWrapper, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	wrapper := new(dataWrapper)
	if err := json.Unmarshal(data, wrapper); err != nil {
		return nil, err
	}

	return wrapper, nil
}

// fakeClient is close to the simplest fake client implementation usable
// by the cache. It does not use pagination like Azure clients.
type fakeClient struct {
	skus []*armcompute.ResourceSKU
	err  error
}

var _ client = &fakeClient{}

func (f *fakeClient) List(ctx context.Context, filter, includeExtendedLocations string) ([]*armcompute.ResourceSKU, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.skus, nil
}

// fakeResourceSKUsClient is a fake client for the real Azure types.
type fakeResourceSKUsClient struct {
	skus [][]*armcompute.ResourceSKU
	err  error
}

var _ ResourceSKUsClient = &fakeResourceSKUsClient{}

func (f *fakeResourceSKUsClient) NewListPager(options *armcompute.ResourceSKUsClientListOptions) *runtime.Pager[armcompute.ResourceSKUsClientListResponse] {
	pageCount := 0
	pager := runtime.NewPager(runtime.PagingHandler[armcompute.ResourceSKUsClientListResponse]{
		More: func(current armcompute.ResourceSKUsClientListResponse) bool {
			return pageCount < len(f.skus)
		},
		Fetcher: func(ctx context.Context, current *armcompute.ResourceSKUsClientListResponse) (armcompute.ResourceSKUsClientListResponse, error) {
			if pageCount >= len(f.skus) {
				return armcompute.ResourceSKUsClientListResponse{}, f.err
			}
			pageCount += 1
			return armcompute.ResourceSKUsClientListResponse{
				ResourceSKUsResult: armcompute.ResourceSKUsResult{
					Value: f.skus[pageCount-1],
				},
			}, f.err
		},
	})
	return pager
}

// newSuccessfulFakeResourceSKUsClient takes a list of sku lists and returns a ResourceSKUsClient.
func newSuccessfulFakeResourceSKUsClient(skuLists [][]*armcompute.ResourceSKU) *fakeResourceSKUsClient {
	return &fakeResourceSKUsClient{
		skus: skuLists,
		err:  nil,
	}
}

// chunk divides a list into count pieces.
func chunk(skus []*armcompute.ResourceSKU, count int) [][]*armcompute.ResourceSKU {
	divided := [][]*armcompute.ResourceSKU{}
	size := (len(skus) + count - 1) / count
	for i := 0; i < len(skus); i += size {
		end := i + size

		if end > len(skus) {
			end = len(skus)
		}

		divided = append(divided, skus[i:end])
	}
	return divided
}
