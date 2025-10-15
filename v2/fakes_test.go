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

// fakeResourceClient is a fake client for the real Azure types. It
// returns a result iterator and can test against arbitrary sequences of
// return pages, injecting failure.
type fakeResourceClient struct {
	skuLists [][]*armcompute.ResourceSKU
	err      error
}

func (f *fakeResourceClient) NewListPager(options *armcompute.ResourceSKUsClientListOptions) *runtime.Pager[armcompute.ResourceSKUsClientListResponse] {
	pageCount := 0
	pager := runtime.NewPager(runtime.PagingHandler[armcompute.ResourceSKUsClientListResponse]{
		More: func(current armcompute.ResourceSKUsClientListResponse) bool {
			return pageCount < len(f.skuLists)
		},
		Fetcher: func(ctx context.Context, current *armcompute.ResourceSKUsClientListResponse) (armcompute.ResourceSKUsClientListResponse, error) {
			if f.err != nil {
				return armcompute.ResourceSKUsClientListResponse{}, f.err
			}
			if pageCount >= len(f.skuLists) {
				return armcompute.ResourceSKUsClientListResponse{}, nil
			}
			pageCount += 1
			return armcompute.ResourceSKUsClientListResponse{
				ResourceSKUsResult: armcompute.ResourceSKUsResult{
					Value: f.skuLists[pageCount-1],
				},
			}, nil
		},
	})
	return pager
}

//nolint:deadcode,unused
func newFailingFakeResourceClient(reterr error) *fakeResourceClient {
	return &fakeResourceClient{
		skuLists: [][]*armcompute.ResourceSKU{{}},
		err:      reterr,
	}
}

// newSuccessfulFakeResourceClient takes a list of sku lists and returns
// a ResourceClient which iterates over all of them, mapping each sku
// list to a page of values.
func newSuccessfulFakeResourceClient(skuLists [][]*armcompute.ResourceSKU) *fakeResourceClient {
	return &fakeResourceClient{
		skuLists: skuLists,
		err:      nil,
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
