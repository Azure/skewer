package skewer

import (
	"context"
	"encoding/json"
	"os"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
)

// dataWrapper is a convenience wrapper for deserializing json testdata
type dataWrapper struct {
	Value []compute.ResourceSku `json:"value,omitempty"`
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
	skus []compute.ResourceSku
	err  error
}

func (f *fakeClient) List(ctx context.Context, filter string) ([]compute.ResourceSku, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.skus, nil
}

// fakeResourceClient is a fake client for the real Azure types. It
// returns a result iterator and can test against arbitrary sequences of
// return pages, injecting failure.
type fakeResourceClient struct {
	res compute.ResourceSkusResultIterator
	err error
}

func (f *fakeResourceClient) ListComplete(ctx context.Context, filter string) (compute.ResourceSkusResultIterator, error) {
	if f.err != nil {
		return compute.ResourceSkusResultIterator{}, f.err
	}
	return f.res, nil
}

// nolint:deadcode,unused
func newFailingFakeResourceClient(reterr error) *fakeResourceClient {
	return &fakeResourceClient{
		res: compute.ResourceSkusResultIterator{},
		err: reterr,
	}
}

// newSuccessfulFakeResourceClient takes a list of sku lists and returns
// a ResourceClient which iterates over all of them, mapping each sku
// list to a page of values.
func newSuccessfulFakeResourceClient(skuLists [][]compute.ResourceSku) (*fakeResourceClient, error) {
	iterator, err := newFakeResourceSkusResultIterator(skuLists)
	if err != nil {
		return nil, err
	}

	return &fakeResourceClient{
		res: iterator,
		err: nil,
	}, nil
}

// fakeResourceProviderClient is a fake client for the real Azure types. It
// returns a result iterator and can test against arbitrary sequences of
// return pages, injecting failure. This uses the resource provider
// signature for testing purposes.
type fakeResourceProviderClient struct {
	res compute.ResourceSkusResultPage
	err error
}

func (f *fakeResourceProviderClient) List(ctx context.Context, filter string) (compute.ResourceSkusResultPage, error) {
	if f.err != nil {
		return compute.ResourceSkusResultPage{}, f.err
	}
	return f.res, nil
}

// nolint:deadcode,unused
func newFailingFakeResourceProviderClient(reterr error) *fakeResourceProviderClient {
	return &fakeResourceProviderClient{
		res: compute.ResourceSkusResultPage{},
		err: reterr,
	}
}

// newSuccessfulFakeResourceProviderClient takes a list of sku lists and returns
// a ResourceProviderClient which iterates over all of them, mapping each sku
// list to a page of values.
func newSuccessfulFakeResourceProviderClient(skuLists [][]compute.ResourceSku) (*fakeResourceProviderClient, error) {
	page, err := newFakeResourceSkusResultPage(skuLists)
	if err != nil {
		return nil, err
	}

	return &fakeResourceProviderClient{
		res: page,
		err: nil,
	}, nil
}

// newFakeResourceSkusResultPage takes a list of sku lists and
// returns an iterator over all items, mapping each sku
// list to a page of values.
func newFakeResourceSkusResultPage(skuLists [][]compute.ResourceSku) (compute.ResourceSkusResultPage, error) {
	pages := newPageList(skuLists)
	newPage := compute.NewResourceSkusResultPage(pages.next)
	if err := newPage.NextWithContext(context.Background()); err != nil {
		return compute.ResourceSkusResultPage{}, err
	}
	return newPage, nil
}

// newFakeResourceSkusResultIterator takes a list of sku lists and
// returns an iterator over all items, mapping each sku
// list to a page of values.
func newFakeResourceSkusResultIterator(skuLists [][]compute.ResourceSku) (compute.ResourceSkusResultIterator, error) {
	pages := newPageList(skuLists)
	newPage := compute.NewResourceSkusResultPage(pages.next)
	if err := newPage.NextWithContext(context.Background()); err != nil {
		return compute.ResourceSkusResultIterator{}, err
	}
	return compute.NewResourceSkusResultIterator(newPage), nil
}

// chunk divides a list into count pieces.
func chunk(skus []compute.ResourceSku, count int) [][]compute.ResourceSku {
	divided := [][]compute.ResourceSku{}
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

// pageList is a utility type to help construct ResourceSkusResultIterators.
type pageList struct {
	cursor int
	pages  []compute.ResourceSkusResult
}

func newPageList(skuLists [][]compute.ResourceSku) *pageList {
	list := &pageList{}
	for i := 0; i < len(skuLists); i++ {
		list.pages = append(list.pages, compute.ResourceSkusResult{
			Value: &skuLists[i],
		})
	}
	return list
}

// next underpins ResourceSkusResultIterator's NextWithDone() method.
func (p *pageList) next(context.Context, compute.ResourceSkusResult) (compute.ResourceSkusResult, error) {
	if p.cursor >= len(p.pages) {
		return compute.ResourceSkusResult{}, nil
	}
	old := p.cursor
	p.cursor++
	return p.pages[old], nil
}
