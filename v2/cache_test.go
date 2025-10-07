package skewer

import (
	"context"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
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
					location: "foo",
				},
			},
		},
	}

	for name, tc := range cases {
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
		unfiltered []*armcompute.ResourceSKU
		condition  FilterFn
		expected   []*armcompute.ResourceSKU
	}{
		"nil slice filters to nil slice": {
			condition: func(*SKU) bool { return true },
		},
		"empty slice filters to empty slice": {
			unfiltered: []*armcompute.ResourceSKU{},
			condition:  func(*SKU) bool { return true },
			expected:   []*armcompute.ResourceSKU{},
		},
		"slice with non-matching element filters to empty slice": {
			unfiltered: []*armcompute.ResourceSKU{
				{
					ResourceType: to.Ptr("nomatch"),
				},
			},
			condition: func(s *SKU) bool { return s.GetName() == "match" },
			expected:  []*armcompute.ResourceSKU{},
		},
		"slice with one matching element doesn't change": {
			unfiltered: []*armcompute.ResourceSKU{
				{
					ResourceType: to.Ptr("match"),
				},
			},
			condition: func(s *SKU) bool { return true },
			expected: []*armcompute.ResourceSKU{
				{
					ResourceType: to.Ptr("match"),
				},
			},
		},
		"all matching elements removed": {
			unfiltered: []*armcompute.ResourceSKU{
				{
					ResourceType: to.Ptr("match"),
				},
				{
					ResourceType: to.Ptr("nomatch"),
				},
				{
					ResourceType: to.Ptr("match"),
				},
				{
					ResourceType: to.Ptr("unmatch"),
				},
				{
					ResourceType: to.Ptr("match"),
				},
			},
			condition: func(s *SKU) bool { return !s.IsResourceType("match") },
			expected: []*armcompute.ResourceSKU{
				{
					ResourceType: to.Ptr("nomatch"),
				},
				{
					ResourceType: to.Ptr("unmatch"),
				},
			},
		},
	}
	for name, tc := range cases {
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

func Test_Cache_Get(t *testing.T) { //nolint:funlen
	cases := map[string]struct {
		sku          string
		resourceType string
		have         []*armcompute.ResourceSKU
		found        bool
	}{
		"should return false with no data": {
			sku:          "foo",
			resourceType: "bar",
		},
		"should match when found at index=0": {
			sku:          "foo",
			resourceType: "bar",
			have: []*armcompute.ResourceSKU{
				{
					Name:         to.Ptr("foo"),
					ResourceType: to.Ptr("bar"),
					Locations:    []*string{to.Ptr("")},
				},
			},
			found: true,
		},
		"should match when found at index=1": {
			sku:          "foo",
			resourceType: "bar",
			have: []*armcompute.ResourceSKU{
				{
					Name:         to.Ptr("other"),
					ResourceType: to.Ptr("baz"),
				},
				{
					Name:         to.Ptr("foo"),
					ResourceType: to.Ptr("bar"),
					Locations:    []*string{to.Ptr("")},
				},
			},
			found: true,
		},
		"should match regardless of sku capitalization": {
			sku:          "foo",
			resourceType: "bar",
			have: []*armcompute.ResourceSKU{
				{
					Name:         to.Ptr("other"),
					ResourceType: to.Ptr("baz"),
				},
				{
					Name:         to.Ptr("FoO"),
					ResourceType: to.Ptr("bar"),
					Locations:    []*string{to.Ptr("")},
				},
			},
			found: true,
		},
		"should return false when no match exists": {
			sku:          "foo",
			resourceType: "bar",
			have: []*armcompute.ResourceSKU{
				{
					Name: to.Ptr("other"),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cache := &Cache{
				data: Wrap(tc.have),
			}

			val, err := cache.Get(context.Background(), tc.sku, tc.resourceType, "")
			if tc.found {
				if err != nil {
					t.Errorf("expected success when trying to Get resource with name %s and resourceType %s, but got error: '%s'",
						tc.sku,
						tc.resourceType,
						err,
					)
				}
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
			} else if err == nil {
				t.Errorf("expected Get to fail with name %s and resourceType %s, but succeeded", tc.sku, tc.resourceType)
			}
		})
	}
}

func Test_Cache_GetAvailabilityZones(t *testing.T) { //nolint:funlen
	cases := map[string]struct {
		have []*armcompute.ResourceSKU
		want []string
	}{
		"should find 1 result": {
			have: []*armcompute.ResourceSKU{
				{
					Name:         to.Ptr("foo"),
					ResourceType: to.Ptr(string(VirtualMachines)),
					Locations: []*string{
						to.Ptr("baz"),
					},
					LocationInfo: []*armcompute.ResourceSKULocationInfo{
						{
							Location: to.Ptr("baz"),
							Zones:    []*string{to.Ptr("1")},
						},
					},
				},
			},
			want: []string{"1"},
		},
		"should find 2 results": {
			have: []*armcompute.ResourceSKU{
				{
					Name:         to.Ptr("foo"),
					ResourceType: to.Ptr(string(VirtualMachines)),
					Locations: []*string{
						to.Ptr("baz"),
					},
					LocationInfo: []*armcompute.ResourceSKULocationInfo{
						{
							Location: to.Ptr("baz"),
							Zones:    []*string{to.Ptr("1")},
						},
					},
				},
				{
					Name:         to.Ptr("foo"),
					ResourceType: to.Ptr(string(VirtualMachines)),
					Locations: []*string{
						to.Ptr("baz"),
					},
					LocationInfo: []*armcompute.ResourceSKULocationInfo{
						{
							Location: to.Ptr("baz"),
							Zones:    []*string{to.Ptr("2")},
						},
					},
				},
			},
			want: []string{"1", "2"},
		},
		"should not find due to location mismatch": {
			have: []*armcompute.ResourceSKU{
				{
					Name:         to.Ptr("foo"),
					ResourceType: to.Ptr(string(VirtualMachines)),
					Locations: []*string{
						to.Ptr("foobar"),
					},
					LocationInfo: []*armcompute.ResourceSKULocationInfo{
						{
							Location: to.Ptr("foobar"),
							Zones:    []*string{to.Ptr("1")},
						},
					},
				},
			},
			want: nil,
		},
		"should not find due to location restriction": {
			have: []*armcompute.ResourceSKU{
				{
					Name:         to.Ptr("foo"),
					ResourceType: to.Ptr(string(VirtualMachines)),
					Locations: []*string{
						to.Ptr("baz"),
					},
					LocationInfo: []*armcompute.ResourceSKULocationInfo{
						{
							Location: to.Ptr("baz"),
							Zones:    []*string{to.Ptr("1")},
						},
					},
					Restrictions: []*armcompute.ResourceSKURestrictions{
						{
							Type:   to.Ptr(armcompute.ResourceSKURestrictionsTypeLocation),
							Values: []*string{to.Ptr("baz")},
						},
					},
				},
			},
			want: nil,
		},
		"should not find due to zone restriction": {
			have: []*armcompute.ResourceSKU{
				{
					Name:         to.Ptr("foo"),
					ResourceType: to.Ptr(string(VirtualMachines)),
					Locations: []*string{
						to.Ptr("baz"),
					},
					LocationInfo: []*armcompute.ResourceSKULocationInfo{
						{
							Location: to.Ptr("baz"),
							Zones:    []*string{to.Ptr("1")},
						},
					},
					Restrictions: []*armcompute.ResourceSKURestrictions{
						{
							Type:   to.Ptr(armcompute.ResourceSKURestrictionsTypeZone),
							Values: []*string{to.Ptr("baz")},
							RestrictionInfo: &armcompute.ResourceSKURestrictionInfo{
								Zones: []*string{to.Ptr("1")},
							},
						},
					},
				},
			},
			want: nil,
		},
	}

	for name, tc := range cases {
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
				t.Error(diff)
			}
		})
	}
}

func Test_Cache_GetVirtualMachineAvailabilityZonesForSize(t *testing.T) { //nolint:funlen
	cases := map[string]struct {
		have []*armcompute.ResourceSKU
		want []string
	}{
		"should find 1 result": {
			have: []*armcompute.ResourceSKU{
				{
					Name:         to.Ptr("foo"),
					ResourceType: to.Ptr(string(VirtualMachines)),
					Locations: []*string{
						to.Ptr("baz"),
					},
					LocationInfo: []*armcompute.ResourceSKULocationInfo{
						{
							Location: to.Ptr("baz"),
							Zones:    []*string{to.Ptr("1")},
						},
					},
				},
			},
			want: []string{"1"},
		},
		"should find 2 results": {
			have: []*armcompute.ResourceSKU{
				{
					Name:         to.Ptr("foo"),
					ResourceType: to.Ptr(string(VirtualMachines)),
					Locations: []*string{
						to.Ptr("baz"),
					},
					LocationInfo: []*armcompute.ResourceSKULocationInfo{
						{
							Location: to.Ptr("baz"),
							Zones:    []*string{to.Ptr("1"), to.Ptr("2")},
						},
					},
				},
			},
			want: []string{"1", "2"},
		},
		"should not find due to size mismatch": {
			have: []*armcompute.ResourceSKU{
				{
					Name:         to.Ptr("foobar"),
					ResourceType: to.Ptr(string(VirtualMachines)),
					Locations: []*string{
						to.Ptr("baz"),
					},
					LocationInfo: []*armcompute.ResourceSKULocationInfo{
						{
							Location: to.Ptr("baz"),
							Zones:    []*string{to.Ptr("1")},
						},
					},
				},
			},
			want: nil,
		},
		"should not find due to location mismatch": {
			have: []*armcompute.ResourceSKU{
				{
					Name:         to.Ptr("foo"),
					ResourceType: to.Ptr(string(VirtualMachines)),
					Locations: []*string{
						to.Ptr("foobar"),
					},
					LocationInfo: []*armcompute.ResourceSKULocationInfo{
						{
							Location: to.Ptr("foobar"),
							Zones:    []*string{to.Ptr("1")},
						},
					},
				},
			},
			want: nil,
		},
		"should not find due to location restriction": {
			have: []*armcompute.ResourceSKU{
				{
					Name:         to.Ptr("foo"),
					ResourceType: to.Ptr(string(VirtualMachines)),
					Locations: []*string{
						to.Ptr("baz"),
					},
					LocationInfo: []*armcompute.ResourceSKULocationInfo{
						{
							Location: to.Ptr("baz"),
							Zones:    []*string{to.Ptr("1")},
						},
					},
					Restrictions: []*armcompute.ResourceSKURestrictions{
						{
							Type:   to.Ptr(armcompute.ResourceSKURestrictionsTypeLocation),
							Values: []*string{to.Ptr("baz")},
						},
					},
				},
			},
			want: nil,
		},
		"should not find due to zone restriction": {
			have: []*armcompute.ResourceSKU{
				{
					Name:         to.Ptr("foo"),
					ResourceType: to.Ptr(string(VirtualMachines)),
					Locations: []*string{
						to.Ptr("baz"),
					},
					LocationInfo: []*armcompute.ResourceSKULocationInfo{
						{
							Location: to.Ptr("baz"),
							Zones:    []*string{to.Ptr("1")},
						},
					},
					Restrictions: []*armcompute.ResourceSKURestrictions{
						{
							Type:   to.Ptr(armcompute.ResourceSKURestrictionsTypeZone),
							Values: []*string{to.Ptr("baz")},
							RestrictionInfo: &armcompute.ResourceSKURestrictionInfo{
								Zones: []*string{to.Ptr("1")},
							},
						},
					},
				},
			},
			want: nil,
		},
	}

	for name, tc := range cases {
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
				t.Fatal(diff)
			}
		})
	}
}
