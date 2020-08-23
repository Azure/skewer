package skewer

import (
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/go-cmp/cmp"
)

func Test_SKU_GetCapabilityQuantity(t *testing.T) {
	cases := map[string]struct {
		sku        compute.ResourceSku
		capability string
		expect     int64
		err        string
	}{
		"empty capability list should return capability not found": {
			sku:        compute.ResourceSku{},
			capability: "",
			err:        (&ErrCapabilityNotFound{""}).Error(),
		},
		"empty capability should not match sku with empty list of capabilities": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{},
			},
			capability: "",
			err:        (&ErrCapabilityNotFound{""}).Error(),
		},
		"empty capability should fail to parse when not integer": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr(""),
						Value: to.StringPtr("False"),
					},
				},
			},
			capability: "",
			err:        "CapabilityValueParse: failed to parse string 'False' as int64, error: 'strconv.ParseInt: parsing \"False\": invalid syntax'", // nolint:lll
		},
		"foo capability should return successfully with integer": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr("foo"),
						Value: to.StringPtr("100"),
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
		sku        compute.ResourceSku
		capability string
		expect     bool
	}{
		"empty capability should not match empty sku": {
			sku:        compute.ResourceSku{},
			capability: "",
		},
		"empty capability should not match sku with empty list of capabilities": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{},
			},
			capability: "",
		},
		"empty capability should not match when present and false": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr(""),
						Value: to.StringPtr("False"),
					},
				},
			},
			capability: "",
		},
		"empty capability should not match when present and weird value": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr(""),
						Value: to.StringPtr("foobar"),
					},
				},
			},
			capability: "",
		},
		"foo capability should not match when false": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr("foo"),
						Value: to.StringPtr("False"),
					},
				},
			},
			capability: "foo",
		},
		"foo capability should match when true": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr("foo"),
						Value: to.StringPtr("True"),
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

func Test_SKU_HasCapabilityWithCapacity(t *testing.T) {
	cases := map[string]struct {
		sku        compute.ResourceSku
		capability string
		capacity   int64
		expect     bool
		err        error
	}{
		"empty capability should not match empty sku": {
			sku:        compute.ResourceSku{},
			capability: "",
		},
		"empty capability should not match sku with empty list of capabilities": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{},
			},
			capability: "",
		},
		"empty capability should error when present and weird value": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr(""),
						Value: to.StringPtr("foobar"),
					},
				},
			},
			capability: "",
			err:        fmt.Errorf("failed to parse string 'foobar' as int64: strconv.ParseInt: parsing \"foobar\": invalid syntax"),
		},
		"empty capability should  match when present with zero capacity and requesting zero": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr(""),
						Value: to.StringPtr("0"),
					},
				},
			},
			capability: "",
			expect:     true,
		},
		"foo capability should not match when present and less than capacity": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr("foo"),
						Value: to.StringPtr("100"),
					},
				},
			},
			capability: "foo",
			capacity:   200,
		},
		"foo capability should match when true": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr("foo"),
						Value: to.StringPtr("10"),
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
			got, err := sku.HasCapabilityWithCapacity(tc.capability, tc.capacity)
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
		sku                compute.ResourceSku
		expectName         string
		expectResourceType string
	}{
		"nil resourceType should return empty string": {
			sku:                compute.ResourceSku{},
			expectResourceType: "",
			expectName:         "",
		},
		"empty resourceType should return empty string": {
			sku: compute.ResourceSku{
				Name:         to.StringPtr(""),
				ResourceType: to.StringPtr(""),
			},
			expectResourceType: "",
			expectName:         "",
		},
		"populated resourceType should return correctly": {
			sku: compute.ResourceSku{
				Name:         to.StringPtr("foo"),
				ResourceType: to.StringPtr("foo"),
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
		sku          compute.ResourceSku
		resourceType string
		expect       bool
	}{
		"nil resourceType should not match anything": {
			sku:          compute.ResourceSku{},
			resourceType: "",
		},
		"empty resourceType should match empty string": {
			sku: compute.ResourceSku{
				ResourceType: to.StringPtr(""),
			},
			resourceType: "",
			expect:       true,
		},
		"empty resourceType should not match non-empty string": {
			sku: compute.ResourceSku{
				ResourceType: to.StringPtr(""),
			},
			resourceType: "foo",
		},
		"populated resourceType should match itself": {
			sku: compute.ResourceSku{
				ResourceType: to.StringPtr("foo"),
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
		sku    compute.ResourceSku
		expect string
	}{
		"nil locations should return empty string": {
			sku:    compute.ResourceSku{},
			expect: "",
		},
		"empty array of locations return empty string": {
			sku: compute.ResourceSku{
				Locations: &[]string{},
			},
			expect: "",
		},
		"single empty value should return empty string": {
			sku: compute.ResourceSku{
				Locations: &[]string{
					"",
				},
			},
			expect: "",
		},
		"populated location should return correctly": {
			sku: compute.ResourceSku{
				Locations: &[]string{
					"foo",
				},
			},
			expect: "foo",
		},
		"should return first with multiple choices (sku api prefers to prefer 1 ResourceSku per location, synthetic test scenario only)": {
			sku: compute.ResourceSku{
				Locations: &[]string{
					"bar",
					"foo",
				},
			},
			expect: "bar",
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			sku := SKU(tc.sku)
			if diff := cmp.Diff(tc.expect, sku.GetLocation()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func Test_SKU_AvailabilityZones(t *testing.T) {}
