package skewer

import (
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
	"github.com/google/go-cmp/cmp"
)

func Test_SKU_GetCapabilityQuantity(t *testing.T) {
	cases := map[string]struct {
		sku        armcompute.ResourceSKU
		capability string
		expect     int64
		err        string
	}{
		"empty capability list should return capability not found": {
			sku:        armcompute.ResourceSKU{},
			capability: "",
			err:        (&ErrCapabilityNotFound{""}).Error(),
		},
		"empty capability should not match sku with empty list of capabilities": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{},
			},
			capability: "",
			err:        (&ErrCapabilityNotFound{""}).Error(),
		},
		"empty capability should fail to parse when not integer": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr(""),
						Value: to.Ptr("False"),
					},
				},
			},
			capability: "",
			err:        "CapabilityValueParse: failed to parse string 'False' as int64, error: 'strconv.ParseInt: parsing \"False\": invalid syntax'", //nolint:lll
		},
		"foo capability should return successfully with integer": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr("foo"),
						Value: to.Ptr("100"),
					},
				},
			},
			capability: "foo",
			expect:     100,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			sku := SKU(tc.sku)
			quantity, err := sku.GetCapabilityIntegerQuantity(tc.capability)
			if tc.err != "" {
				if err == nil {
					t.Errorf("expected failure with error '%s' but did not occur", tc.err)
				}
				if diff := cmp.Diff(tc.err, err.Error()); diff != "" {
					t.Error(diff)
				}
			} else {
				if err != nil {
					t.Errorf("expected success but failure occurred with error '%s'", err)
				}
				if diff := cmp.Diff(tc.expect, quantity); diff != "" {
					t.Error(diff)
				}
			}
		})
	}
}

func Test_SKU_HasCapability(t *testing.T) {
	cases := map[string]struct {
		sku        armcompute.ResourceSKU
		capability string
		expect     bool
	}{
		"empty capability should not match empty sku": {
			sku:        armcompute.ResourceSKU{},
			capability: "",
		},
		"empty capability should not match sku with empty list of capabilities": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{},
			},
			capability: "",
		},
		"empty capability should not match when present and false": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr(""),
						Value: to.Ptr("False"),
					},
				},
			},
			capability: "",
		},
		"empty capability should not match when present and weird value": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr(""),
						Value: to.Ptr("foobar"),
					},
				},
			},
			capability: "",
		},
		"foo capability should not match when false": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr("foo"),
						Value: to.Ptr("False"),
					},
				},
			},
			capability: "foo",
		},
		"foo capability should match when true": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr("foo"),
						Value: to.Ptr("True"),
					},
				},
			},
			capability: "foo",
			expect:     true,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			sku := SKU(tc.sku)
			if diff := cmp.Diff(tc.expect, sku.HasCapability(tc.capability)); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func Test_SKU_HasCapabilityWithMinCapacity(t *testing.T) {
	cases := map[string]struct {
		sku        armcompute.ResourceSKU
		capability string
		capacity   int64
		expect     bool
		err        error
	}{
		"empty capability should not match empty sku": {
			sku:        armcompute.ResourceSKU{},
			capability: "",
		},
		"empty capability should not match sku with empty list of capabilities": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{},
			},
			capability: "",
		},
		"empty capability should error when present and weird value": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr(""),
						Value: to.Ptr("foobar"),
					},
				},
			},
			capability: "",
			err:        fmt.Errorf("failed to parse string 'foobar' as int64: strconv.ParseInt: parsing \"foobar\": invalid syntax"),
		},
		"empty capability should  match when present with zero capacity and requesting zero": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr(""),
						Value: to.Ptr("0"),
					},
				},
			},
			capability: "",
			expect:     true,
		},
		"foo capability should not match when present and less than capacity": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr("foo"),
						Value: to.Ptr("100"),
					},
				},
			},
			capability: "foo",
			capacity:   200,
		},
		"foo capability should match when true": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr("foo"),
						Value: to.Ptr("10"),
					},
				},
			},
			capability: "foo",
			capacity:   5,
			expect:     true,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			sku := SKU(tc.sku)
			got, err := sku.HasCapabilityWithMinCapacity(tc.capability, tc.capacity)
			if tc.err != nil {
				if diff := cmp.Diff(tc.err.Error(), err.Error()); diff != "" {
					t.Error(diff)
				}
			} else {
				if diff := cmp.Diff(tc.expect, got); diff != "" {
					t.Error(diff)
				}
			}
		})
	}
}

func Test_SKU_GetResourceTypeAndName(t *testing.T) {
	cases := map[string]struct {
		sku                armcompute.ResourceSKU
		expectName         string
		expectResourceType string
	}{
		"nil resourceType should return empty string": {
			sku:                armcompute.ResourceSKU{},
			expectResourceType: "",
			expectName:         "",
		},
		"empty resourceType should return empty string": {
			sku: armcompute.ResourceSKU{
				Name:         to.Ptr(""),
				ResourceType: to.Ptr(""),
			},
			expectResourceType: "",
			expectName:         "",
		},
		"populated resourceType should return correctly": {
			sku: armcompute.ResourceSKU{
				Name:         to.Ptr("foo"),
				ResourceType: to.Ptr("foo"),
			},
			expectResourceType: "foo",
			expectName:         "foo",
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			sku := SKU(tc.sku)
			if diff := cmp.Diff(tc.expectName, sku.GetName()); diff != "" {
				t.Errorf("mismatched name\n%s", diff)
			}
			if diff := cmp.Diff(tc.expectResourceType, sku.GetResourceType()); diff != "" {
				t.Errorf("mismatched resourceType\n%s", diff)
			}
		})
	}
}

func Test_SKU_IsResourceType(t *testing.T) {
	cases := map[string]struct {
		sku          armcompute.ResourceSKU
		resourceType string
		expect       bool
	}{
		"nil resourceType should not match anything": {
			sku:          armcompute.ResourceSKU{},
			resourceType: "",
		},
		"empty resourceType should match empty string": {
			sku: armcompute.ResourceSKU{
				ResourceType: to.Ptr(""),
			},
			resourceType: "",
			expect:       true,
		},
		"empty resourceType should not match non-empty string": {
			sku: armcompute.ResourceSKU{
				ResourceType: to.Ptr(""),
			},
			resourceType: "foo",
		},
		"populated resourceType should match itself": {
			sku: armcompute.ResourceSKU{
				ResourceType: to.Ptr("foo"),
			},
			resourceType: "foo",
			expect:       true,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			sku := SKU(tc.sku)
			if diff := cmp.Diff(tc.expect, sku.IsResourceType(tc.resourceType)); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func Test_SKU_GetLocation(t *testing.T) {
	cases := map[string]struct {
		sku       armcompute.ResourceSKU
		expect    string
		expectErr string
	}{
		"nil locations should return empty string": {
			sku:    armcompute.ResourceSKU{},
			expect: "",
		},
		"empty array of locations return empty string": {
			sku: armcompute.ResourceSKU{
				Locations: []*string{},
			},
			expect: "",
		},
		"single empty value should return empty string": {
			sku: armcompute.ResourceSKU{
				Locations: []*string{
					to.Ptr(""),
				},
			},
			expect: "",
		},
		"populated location should return correctly": {
			sku: armcompute.ResourceSKU{
				Locations: []*string{
					to.Ptr("foo"),
				},
			},
			expect: "foo",
		},
		"should return error with multiple choices": {
			sku: armcompute.ResourceSKU{
				Locations: []*string{
					to.Ptr("bar"),
					to.Ptr("foo"),
				},
			},
			expectErr: "sku had multiple locations, refusing to disambiguate",
		},
		"should return error with no choices": {
			sku: armcompute.ResourceSKU{
				Locations: []*string{},
			},
			expectErr: "sku had no locations",
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			sku := SKU(tc.sku)
			got, err := sku.GetLocation()
			if tc.expectErr != "" {
				if err == nil {
					t.Errorf("expected error '%s', but got none", tc.expectErr)
				}
				if err.Error() != tc.expectErr {
					t.Errorf("expected error '%s', but got '%s'", tc.expectErr, err.Error())
				}
			}
			if diff := cmp.Diff(tc.expect, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func Test_SKU_AvailabilityZones(t *testing.T) {}

//nolint:funlen
func Test_SKU_HasCapabilityInZone(t *testing.T) {
	cases := map[string]struct {
		sku        armcompute.ResourceSKU
		capability string
		zone       string
		expect     bool
	}{
		"should return false when capability is false": {
			sku: armcompute.ResourceSKU{
				LocationInfo: []*armcompute.ResourceSKULocationInfo{
					{
						ZoneDetails: []*armcompute.ResourceSKUZoneDetails{
							{
								Name: []*string{to.Ptr("1"), to.Ptr("3")},
								Capabilities: []*armcompute.ResourceSKUCapabilities{
									{
										Name:  to.Ptr("foo"),
										Value: to.Ptr("False"),
									},
								},
							},
						},
					},
				},
			},
			capability: "foo",
			zone:       "1",
			expect:     false,
		},
		"should return false when zone doesn't match": {
			sku: armcompute.ResourceSKU{
				LocationInfo: []*armcompute.ResourceSKULocationInfo{
					{
						ZoneDetails: []*armcompute.ResourceSKUZoneDetails{
							{
								Name: []*string{to.Ptr("1"), to.Ptr("3")},
								Capabilities: []*armcompute.ResourceSKUCapabilities{
									{
										Name:  to.Ptr("foo"),
										Value: to.Ptr("True"),
									},
								},
							},
						},
					},
				},
			},
			capability: "foo",
			zone:       "2",
			expect:     false,
		},
		"should not return true when the capability is not set in availability zone but set on sku capability": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr("foo"),
						Value: to.Ptr("True"),
					},
				},
			},
			capability: "foo",
			zone:       "1",
			expect:     false,
		},
		"should return true when capability and zone match": {
			sku: armcompute.ResourceSKU{
				LocationInfo: []*armcompute.ResourceSKULocationInfo{
					{
						ZoneDetails: []*armcompute.ResourceSKUZoneDetails{
							{
								Name: []*string{to.Ptr("1"), to.Ptr("3")},
								Capabilities: []*armcompute.ResourceSKUCapabilities{
									{
										Name:  to.Ptr("foo"),
										Value: to.Ptr("True"),
									},
								},
							},
						},
					},
				},
			},
			capability: "foo",
			zone:       "1",
			expect:     true,
		},
		"should return true when capability and zone match for zone 3": {
			sku: armcompute.ResourceSKU{
				LocationInfo: []*armcompute.ResourceSKULocationInfo{
					{
						ZoneDetails: []*armcompute.ResourceSKUZoneDetails{
							{
								Name: []*string{to.Ptr("1"), to.Ptr("3")},
								Capabilities: []*armcompute.ResourceSKUCapabilities{
									{
										Name:  to.Ptr("foo"),
										Value: to.Ptr("True"),
									},
								},
							},
						},
					},
				},
			},
			capability: "foo",
			zone:       "3",
			expect:     true,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			sku := SKU(tc.sku)
			if diff := cmp.Diff(tc.expect, sku.HasCapabilityInZone(tc.capability, tc.zone)); diff != "" {
				t.Error(diff)
			}
		})
	}
}

// Test_SKU_MemberOf tests the SKU MemberOf method
func Test_SKU_Includes(t *testing.T) {
	cases := map[string]struct {
		skuList []SKU
		sku     SKU
		expect  bool
	}{
		"empty list should not include": {
			skuList: []SKU{},
			sku: SKU{
				Name: to.Ptr("foo"),
			},
			expect: false,
		},
		"missing name should not include": {
			skuList: []SKU{
				{
					Name: to.Ptr("foo"),
				},
			},
			sku: SKU{
				Name: to.Ptr("bar"),
			},
			expect: false,
		},
		"name is included": {
			skuList: []SKU{
				{
					Name: to.Ptr("foo"),
				},
				{
					Name: to.Ptr("bar"),
				},
			},
			sku: SKU{
				Name: to.Ptr("bar"),
			},
			expect: true,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if diff := cmp.Diff(tc.expect, tc.sku.MemberOf(tc.skuList)); diff != "" {
				t.Error(diff)
			}
		})
	}
}
