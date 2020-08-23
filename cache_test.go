package skewer

import (
	"context"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func Test_NewCache(t *testing.T) {
	cases := map[string]struct{}{}
	_ = cases
}

func Test_WithLocation(t *testing.T) {
	cases := map[string]struct {
		options []Option
		expect  *Cache
	}{
		"should be empty with no options": {
			expect: &Cache{
				config: &Config{},
			},
		},
		"should have location and filter": {
			options: []Option{WithLocation("foo")},
			expect: &Cache{
				config: &Config{
					filter:   "location eq 'foo'",
					Location: "foo",
				},
			},
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			cache, err := NewStaticCache(nil, tc.options...)
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(tc.expect.config, cache.config); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func Test_Cache_List(t *testing.T) {
	cases := map[string]struct{}{}
	_ = cases
}
func Test_Cache_GetVirtualMachines(t *testing.T) {
	cases := map[string]struct{}{}
	_ = cases
}

func Test_Filter(t *testing.T) {
	cases := map[string]struct {
		unfiltered []compute.ResourceSku
		condition  FilterFn
		expected   []compute.ResourceSku
	}{
		"nil slice filters to nil slice": {
			condition: func(*SKU) bool { return true },
		},
		"empty slice filters to empty slice": {
			unfiltered: []compute.ResourceSku{},
			condition:  func(*SKU) bool { return true },
			expected:   []compute.ResourceSku{},
		},
		"slice with non-matching element filters to empty slice": {
			unfiltered: []compute.ResourceSku{
				{
					ResourceType: to.StringPtr("nomatch"),
				},
			},
			condition: func(s *SKU) bool { return s.GetName() == "match" },
			expected:  []compute.ResourceSku{},
		},
		"slice with one matching element doesn't change": {
			unfiltered: []compute.ResourceSku{
				{
					ResourceType: to.StringPtr("match"),
				},
			},
			condition: func(s *SKU) bool { return true },
			expected: []compute.ResourceSku{
				{
					ResourceType: to.StringPtr("match"),
				},
			},
		},
		"all matching elements removed": {
			unfiltered: []compute.ResourceSku{
				{
					ResourceType: to.StringPtr("match"),
				},
				{
					ResourceType: to.StringPtr("nomatch"),
				},
				{
					ResourceType: to.StringPtr("match"),
				},
				{
					ResourceType: to.StringPtr("unmatch"),
				},
				{
					ResourceType: to.StringPtr("match"),
				},
			},
			condition: func(s *SKU) bool { return !s.IsResourceType("match") },
			expected: []compute.ResourceSku{
				{
					ResourceType: to.StringPtr("nomatch"),
				},
				{
					ResourceType: to.StringPtr("unmatch"),
				},
			},
		},
	}
	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			result := Filter(Wrap(tc.unfiltered), tc.condition)
			if diff := cmp.Diff(result, Wrap(tc.expected), []cmp.Option{
				cmpopts.EquateEmpty(),
			}...); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func Test_Map(t *testing.T) {
	t.Run("nil slice maps to nil slice", func(t *testing.T) {
		mapFn := func(*SKU) SKU { return SKU{} }
		if Map(nil, mapFn) != nil {
			t.Error()
		}
	})

	t.Run("empty slice maps to empty slice", func(t *testing.T) {
		mapFn := func(*SKU) SKU { return SKU{} }
		if len(Map([]SKU{}, mapFn)) != 0 {
			t.Error()
		}
	})

	t.Run("identity function keeps slice the same", func(t *testing.T) {
		mapFn := func(s *SKU) SKU { return *s }
		skuList := make([]SKU, 100)
		mapped := Map(skuList, mapFn)
		if diff := cmp.Diff(mapped, skuList); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("map hits each element once", func(t *testing.T) {
		counter := 0
		skuList := make([]SKU, 100)
		Map(skuList, func(s *SKU) SKU {
			counter++
			return SKU{}
		})

		if counter != 100 {
			t.Error()
		}
	})
}

func Test_Cache_Get(t *testing.T) {
	cases := map[string]struct {
		sku          string
		resourceType string
		have         []compute.ResourceSku
		found        bool
	}{
		"should return false with no data": {
			sku:          "foo",
			resourceType: "bar",
		},
		"should match when found at index=0": {
			sku:          "foo",
			resourceType: "bar",
			have: []compute.ResourceSku{
				{
					Name:         to.StringPtr("foo"),
					ResourceType: to.StringPtr("bar"),
				},
			},
			found: true,
		},
		"should match when found at index=1": {
			sku:          "foo",
			resourceType: "bar",
			have: []compute.ResourceSku{
				{
					Name:         to.StringPtr("other"),
					ResourceType: to.StringPtr("baz"),
				},
				{
					Name:         to.StringPtr("foo"),
					ResourceType: to.StringPtr("bar"),
				},
			},
			found: true,
		},
		"should match regardless of sku capitalization": {
			sku:          "foo",
			resourceType: "bar",
			have: []compute.ResourceSku{
				{
					Name:         to.StringPtr("other"),
					ResourceType: to.StringPtr("baz"),
				},
				{
					Name:         to.StringPtr("FoO"),
					ResourceType: to.StringPtr("bar"),
				},
			},
			found: true,
		},
		"should return false when no match exists": {
			sku:          "foo",
			resourceType: "bar",
			have: []compute.ResourceSku{
				{
					Name: to.StringPtr("other"),
				},
			},
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			cache := &Cache{
				data: Wrap(tc.have),
			}

			val, found := cache.Get(context.Background(), tc.sku, tc.resourceType)
			if tc.found != found {
				t.Errorf("expected %t but got %t when trying to Get resource with name %s and resourceType %s",
					tc.found,
					found,
					tc.sku,
					tc.resourceType,
				)
			} else if found {
				if val.Name == nil {
					t.Fatalf("expected name to be %s, but was nil", tc.sku)
					return
				}
				if !strings.EqualFold(*val.Name, tc.sku) {
					t.Fatalf("expected name to be %s, but was %s", tc.sku, *val.Name)
				}
				if val.ResourceType == nil {
					t.Fatalf("expected name to be %s, but was nil", tc.sku)
					return
				}
				if *val.ResourceType != tc.resourceType {
					t.Fatalf("expected kind to be %s, but was %s", tc.resourceType, *val.ResourceType)
				}
			}
		})
	}
}

func Test_Cache_GetAvailabilityZones(t *testing.T) { //nolint:funlen
	cases := map[string]struct {
		have []compute.ResourceSku
		want []string
	}{
		"should find 1 result": {
			have: []compute.ResourceSku{
				{
					Name:         to.StringPtr("foo"),
					ResourceType: to.StringPtr(string(VirtualMachines)),
					Locations: &[]string{
						"baz",
					},
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("baz"),
							Zones:    &[]string{"1"},
						},
					},
				},
			},
			want: []string{"1"},
		},
		"should find 2 results": {
			have: []compute.ResourceSku{
				{
					Name:         to.StringPtr("foo"),
					ResourceType: to.StringPtr(string(VirtualMachines)),
					Locations: &[]string{
						"baz",
					},
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("baz"),
							Zones:    &[]string{"1"},
						},
					},
				},
				{
					Name:         to.StringPtr("foo"),
					ResourceType: to.StringPtr(string(VirtualMachines)),
					Locations: &[]string{
						"baz",
					},
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("baz"),
							Zones:    &[]string{"2"},
						},
					},
				},
			},
			want: []string{"1", "2"},
		},
		"should not find due to location mismatch": {
			have: []compute.ResourceSku{
				{
					Name:         to.StringPtr("foo"),
					ResourceType: to.StringPtr(string(VirtualMachines)),
					Locations: &[]string{
						"foobar",
					},
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("foobar"),
							Zones:    &[]string{"1"},
						},
					},
				},
			},
			want: nil,
		},
		"should not find due to location restriction": {
			have: []compute.ResourceSku{
				{
					Name:         to.StringPtr("foo"),
					ResourceType: to.StringPtr(string(VirtualMachines)),
					Locations: &[]string{
						"baz",
					},
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("baz"),
							Zones:    &[]string{"1"},
						},
					},
					Restrictions: &[]compute.ResourceSkuRestrictions{
						{
							Type:   compute.Location,
							Values: &[]string{"baz"},
						},
					},
				},
			},
			want: nil,
		},
		"should not find due to zone restriction": {
			have: []compute.ResourceSku{
				{
					Name:         to.StringPtr("foo"),
					ResourceType: to.StringPtr(string(VirtualMachines)),
					Locations: &[]string{
						"baz",
					},
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("baz"),
							Zones:    &[]string{"1"},
						},
					},
					Restrictions: &[]compute.ResourceSkuRestrictions{
						{
							Type: compute.Zone,
							RestrictionInfo: &compute.ResourceSkuRestrictionInfo{
								Zones: &[]string{"1"},
							},
						},
					},
				},
			},
			want: nil,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			cache, err := NewStaticCache(Wrap(tc.have), WithLocation("baz"))
			if err != nil {
				t.Error(err)
			}
			zones := cache.GetAvailabilityZones(context.Background())
			if diff := cmp.Diff(zones, tc.want, []cmp.Option{
				cmpopts.EquateEmpty(),
				cmpopts.SortSlices(func(a, b string) bool {
					return a < b
				}),
			}...); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func Test_Cache_GetVirtualMachineAvailabilityZonesForSize(t *testing.T) { //nolint:funlen
	cases := map[string]struct {
		have []compute.ResourceSku
		want []string
	}{
		"should find 1 result": {
			have: []compute.ResourceSku{
				{
					Name:         to.StringPtr("foo"),
					ResourceType: to.StringPtr(string(VirtualMachines)),
					Locations: &[]string{
						"baz",
					},
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("baz"),
							Zones:    &[]string{"1"},
						},
					},
				},
			},
			want: []string{"1"},
		},
		"should find 2 results": {
			have: []compute.ResourceSku{
				{
					Name:         to.StringPtr("foo"),
					ResourceType: to.StringPtr(string(VirtualMachines)),
					Locations: &[]string{
						"baz",
					},
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("baz"),
							Zones:    &[]string{"1", "2"},
						},
					},
				},
			},
			want: []string{"1", "2"},
		},
		"should not find due to size mismatch": {
			have: []compute.ResourceSku{
				{
					Name:         to.StringPtr("foobar"),
					ResourceType: to.StringPtr(string(VirtualMachines)),
					Locations: &[]string{
						"baz",
					},
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("baz"),
							Zones:    &[]string{"1"},
						},
					},
				},
			},
			want: nil,
		},
		"should not find due to location mismatch": {
			have: []compute.ResourceSku{
				{
					Name:         to.StringPtr("foo"),
					ResourceType: to.StringPtr(string(VirtualMachines)),
					Locations: &[]string{
						"foobar",
					},
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("foobar"),
							Zones:    &[]string{"1"},
						},
					},
				},
			},
			want: nil,
		},
		"should not find due to location restriction": {
			have: []compute.ResourceSku{
				{
					Name:         to.StringPtr("foo"),
					ResourceType: to.StringPtr(string(VirtualMachines)),
					Locations: &[]string{
						"baz",
					},
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("baz"),
							Zones:    &[]string{"1"},
						},
					},
					Restrictions: &[]compute.ResourceSkuRestrictions{
						{
							Type:   compute.Location,
							Values: &[]string{"baz"},
						},
					},
				},
			},
			want: nil,
		},
		"should not find due to zone restriction": {
			have: []compute.ResourceSku{
				{
					Name:         to.StringPtr("foo"),
					ResourceType: to.StringPtr(string(VirtualMachines)),
					Locations: &[]string{
						"baz",
					},
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("baz"),
							Zones:    &[]string{"1"},
						},
					},
					Restrictions: &[]compute.ResourceSkuRestrictions{
						{
							Type: compute.Zone,
							RestrictionInfo: &compute.ResourceSkuRestrictionInfo{
								Zones: &[]string{"1"},
							},
						},
					},
				},
			},
			want: nil,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			cache, err := NewStaticCache(Wrap(tc.have), WithLocation("baz"))
			if err != nil {
				t.Error(err)
			}
			zones := cache.GetVirtualMachineAvailabilityZonesForSize(context.Background(), "foo")
			if diff := cmp.Diff(zones, tc.want, []cmp.Option{
				cmpopts.EquateEmpty(),
				cmpopts.SortSlices(func(a, b string) bool {
					return a < b
				}),
			}...); diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}
