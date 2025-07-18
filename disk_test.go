package skewer

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute" //nolint:staticcheck
	"github.com/Azure/go-autorest/autorest/to"
)

func Test_SKU_HasSCSISupport(t *testing.T) {
	cases := map[string]struct {
		sku    compute.ResourceSku
		expect bool
	}{
		"empty capability list should return true (backward compatibility)": {
			sku:    compute.ResourceSku{},
			expect: true,
		},
		"no disk controller capability should return true": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{},
			},
			expect: true,
		},
		"SCSI only should return true": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr(DiskControllerTypes),
						Value: to.StringPtr("SCSI"),
					},
				},
			},
			expect: true,
		},
		"SCSI and NVMe should return true": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr(DiskControllerTypes),
						Value: to.StringPtr("SCSI,NVMe"),
					},
				},
			},
			expect: true,
		},
		"NVMe only should return false": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr(DiskControllerTypes),
						Value: to.StringPtr("NVMe"),
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
		sku    compute.ResourceSku
		expect bool
	}{
		"empty capability list should return false": {
			sku:    compute.ResourceSku{},
			expect: false,
		},
		"no disk controller capability should return false": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{},
			},
			expect: false,
		},
		"SCSI only should return false": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr(DiskControllerTypes),
						Value: to.StringPtr("SCSI"),
					},
				},
			},
			expect: false,
		},
		"SCSI and NVMe should return true": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr(DiskControllerTypes),
						Value: to.StringPtr("SCSI,NVMe"),
					},
				},
			},
			expect: true,
		},
		"NVMe only should return true": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr(DiskControllerTypes),
						Value: to.StringPtr("NVMe"),
					},
				},
			},
			expect: true,
		},
		"NVMe in mixed case should return true": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr(DiskControllerTypes),
						Value: to.StringPtr("SCSI,NVMe,Other"),
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
		sku    compute.ResourceSku
		expect bool
	}{
		"empty capability list should return false": {
			sku:    compute.ResourceSku{},
			expect: false,
		},
		"no ephemeral placement capability should return false": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr("vCPUs"),
						Value: to.StringPtr("8"),
					},
				},
			},
			expect: false,
		},
		"ResourceDisk only should return false": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr(SupportedEphemeralOSDiskPlacements),
						Value: to.StringPtr("ResourceDisk"),
					},
				},
			},
			expect: false,
		},
		"NvmeDisk should return true": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr(SupportedEphemeralOSDiskPlacements),
						Value: to.StringPtr("NvmeDisk"),
					},
				},
			},
			expect: true,
		},
		"ResourceDisk and NvmeDisk should return true": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr(SupportedEphemeralOSDiskPlacements),
						Value: to.StringPtr("ResourceDisk,NvmeDisk"),
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
		sku    compute.ResourceSku
		expect int64
		err    string
	}{
		"empty capability list should return error": {
			sku: compute.ResourceSku{},
			err: (&ErrCapabilityNotFound{NvmeDiskSizeInMiB}).Error(),
		},
		"no NVMe disk size capability should return error": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr("vCPUs"),
						Value: to.StringPtr("8"),
					},
				},
			},
			err: (&ErrCapabilityNotFound{NvmeDiskSizeInMiB}).Error(),
		},
		"valid NVMe disk size should return value": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr(NvmeDiskSizeInMiB),
						Value: to.StringPtr("1024000"),
					},
				},
			},
			expect: 1024000,
		},
		"invalid NVMe disk size should return parse error": {
			sku: compute.ResourceSku{
				Capabilities: &[]compute.ResourceSkuCapabilities{
					{
						Name:  to.StringPtr(NvmeDiskSizeInMiB),
						Value: to.StringPtr("not-a-number"),
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
