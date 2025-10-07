package skewer

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
)

func Test_SKU_HasSCSISupport(t *testing.T) {
	cases := map[string]struct {
		sku    armcompute.ResourceSKU
		expect bool
	}{
		"empty capability list should return true (backward compatibility)": {
			sku:    armcompute.ResourceSKU{},
			expect: true,
		},
		"no disk controller capability should return true": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{},
			},
			expect: true,
		},
		"SCSI only should return true": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr(DiskControllerTypes),
						Value: to.Ptr("SCSI"),
					},
				},
			},
			expect: true,
		},
		"SCSI and NVMe should return true": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr(DiskControllerTypes),
						Value: to.Ptr("SCSI,NVMe"),
					},
				},
			},
			expect: true,
		},
		"NVMe only should return false": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr(DiskControllerTypes),
						Value: to.Ptr("NVMe"),
					},
				},
			},
			expect: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			sku := SKU(tc.sku)
			actual := sku.HasSCSISupport()
			if actual != tc.expect {
				t.Fatalf("expected %v but got %v", tc.expect, actual)
			}
		})
	}
}

func Test_SKU_HasNVMeSupport(t *testing.T) {
	cases := map[string]struct {
		sku    armcompute.ResourceSKU
		expect bool
	}{
		"empty capability list should return false": {
			sku:    armcompute.ResourceSKU{},
			expect: false,
		},
		"no disk controller capability should return false": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{},
			},
			expect: false,
		},
		"SCSI only should return false": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr(DiskControllerTypes),
						Value: to.Ptr("SCSI"),
					},
				},
			},
			expect: false,
		},
		"SCSI and NVMe should return true": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr(DiskControllerTypes),
						Value: to.Ptr("SCSI,NVMe"),
					},
				},
			},
			expect: true,
		},
		"NVMe only should return true": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr(DiskControllerTypes),
						Value: to.Ptr("NVMe"),
					},
				},
			},
			expect: true,
		},
		"NVMe in mixed case should return true": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr(DiskControllerTypes),
						Value: to.Ptr("SCSI,NVMe,Other"),
					},
				},
			},
			expect: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			sku := SKU(tc.sku)
			actual := sku.HasNVMeSupport()
			if actual != tc.expect {
				t.Fatalf("expected %v but got %v", tc.expect, actual)
			}
		})
	}
}

func Test_SKU_SupportsNVMeEphemeralOSDisk(t *testing.T) {
	cases := map[string]struct {
		sku    armcompute.ResourceSKU
		expect bool
	}{
		"empty capability list should return false": {
			sku:    armcompute.ResourceSKU{},
			expect: false,
		},
		"no ephemeral placement capability should return false": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr("vCPUs"),
						Value: to.Ptr("8"),
					},
				},
			},
			expect: false,
		},
		"ResourceDisk only should return false": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr(SupportedEphemeralOSDiskPlacements),
						Value: to.Ptr("ResourceDisk"),
					},
				},
			},
			expect: false,
		},
		"NvmeDisk should return true": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr(SupportedEphemeralOSDiskPlacements),
						Value: to.Ptr("NvmeDisk"),
					},
				},
			},
			expect: true,
		},
		"ResourceDisk and NvmeDisk should return true": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr(SupportedEphemeralOSDiskPlacements),
						Value: to.Ptr("ResourceDisk,NvmeDisk"),
					},
				},
			},
			expect: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			sku := SKU(tc.sku)
			actual := sku.SupportsNVMeEphemeralOSDisk()
			if actual != tc.expect {
				t.Fatalf("expected %v but got %v", tc.expect, actual)
			}
		})
	}
}

func Test_SKU_NVMeDiskSizeInMiB(t *testing.T) {
	cases := map[string]struct {
		sku    armcompute.ResourceSKU
		expect int64
		err    string
	}{
		"empty capability list should return error": {
			sku: armcompute.ResourceSKU{},
			err: (&ErrCapabilityNotFound{NvmeDiskSizeInMiB}).Error(),
		},
		"no NVMe disk size capability should return error": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr("vCPUs"),
						Value: to.Ptr("8"),
					},
				},
			},
			err: (&ErrCapabilityNotFound{NvmeDiskSizeInMiB}).Error(),
		},
		"valid NVMe disk size should return value": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr(NvmeDiskSizeInMiB),
						Value: to.Ptr("1024000"),
					},
				},
			},
			expect: 1024000,
		},
		"invalid NVMe disk size should return parse error": {
			sku: armcompute.ResourceSKU{
				Capabilities: []*armcompute.ResourceSKUCapabilities{
					{
						Name:  to.Ptr(NvmeDiskSizeInMiB),
						Value: to.Ptr("not-a-number"),
					},
				},
			},
			err: "NvmeDiskSizeInMiBCapabilityValueParse: failed to parse string 'not-a-number' as int64, error: 'strconv.ParseInt: parsing \"not-a-number\": invalid syntax'",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			sku := SKU(tc.sku)
			actual, err := sku.NVMeDiskSizeInMiB()
			if tc.err != "" {
				if err == nil || err.Error() != tc.err {
					t.Fatalf("expected error '%s' but got '%v'", tc.err, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if actual != tc.expect {
					t.Fatalf("expected %d but got %d", tc.expect, actual)
				}
			}
		})
	}
}
