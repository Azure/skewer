package skewer

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/pkg/errors"
)

// ResourceClient is the required Azure client interface used to populate skewer's data.
type ResourceClient interface {
	ListComplete(ctx context.Context, filter string) (compute.ResourceSkusResultIterator, error)
}

// ResourceProviderClient is a convenience interface for uses cases
// specific to Azure resource providers.
type ResourceProviderClient interface {
	List(ctx context.Context, filter string) (compute.ResourceSkusResultPage, error)
}

// client defines the internal interface required by the skewer Cache.
// TODO(ace): implement a lazy iterator with caching (and a cursor?)
type client interface {
	List(ctx context.Context, filter string) ([]compute.ResourceSku, error)
}

// wrappedResourceClient defines a wrapper for the typical Azure client
// signature to collect all resource skus from the iterator returned by ListComplete().
type wrappedResourceClient struct {
	client ResourceClient
}

func newWrappedResourceClient(client ResourceClient) *wrappedResourceClient {
	return &wrappedResourceClient{client}
}

// List greedily traverses all returned sku pages
func (w *wrappedResourceClient) List(ctx context.Context, filter string) ([]compute.ResourceSku, error) {
	return iterate(ctx, filter, w.client.ListComplete)
}

// wrappedResourceProviderClient defines a wrapper for the typical Azure client
// signature to collect all resource skus from the iterator returned by
// List(). It only differs from wrappedResourceClient in signature.
type wrappedResourceProviderClient struct {
	client ResourceProviderClient
}

func newWrappedResourceProviderClient(client ResourceProviderClient) *wrappedResourceProviderClient {
	return &wrappedResourceProviderClient{client}
}

func (w *wrappedResourceProviderClient) ListComplete(ctx context.Context, filter string) (compute.ResourceSkusResultIterator, error) {
	page, err := w.client.List(ctx, filter)
	if err != nil {
		return compute.ResourceSkusResultIterator{}, nil
	}
	return compute.NewResourceSkusResultIterator(page), nil
}

type iterFunc func(context.Context, string) (compute.ResourceSkusResultIterator, error)

// iterate invokes fn to get an iterator, then drains it into an array.
func iterate(ctx context.Context, filter string, fn iterFunc) ([]compute.ResourceSku, error) {
	iter, err := fn(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "could not list resource skus")
	}

	var skus []compute.ResourceSku
	for iter.NotDone() {
		skus = append(skus, iter.Value())
		if err := iter.NextWithContext(ctx); err != nil {
			return skus, errors.Wrap(err, "could not iterate resource skus")
		}
	}

	return skus, nil
}
