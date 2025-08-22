package skewer

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	expectedVirtualMachinesCount            = 4
	expectedAvailabilityZones               = []string{"1", "2", "3"}
	shouldNotBePresentCapabilityNotFoundErr = "ShouldNotBePresentCapabilityNotFound"
	premiumIOCapabilityValueParseErr        = "PremiumIOCapabilityValueParse: failed to parse string 'False' as int64, error: 'strconv.ParseInt: parsing \"False\": invalid syntax'" //nolint:lll
	x64ArchType                             = "x64"
)

//nolint:gocyclo,funlen
func Test_Data(t *testing.T) {
	dataWrapper, err := newDataWrapper("./testdata/eastus.json")
	if err != nil {
		t.Error(err)
	}

	fakeClient := &fakeClient{
		skus: dataWrapper.Value,
	}

	resourceSKUsClient, err := newSuccessfulFakeResourceSKUsClient([][]*armcompute.ResourceSKU{
		dataWrapper.Value,
	})
	if err != nil {
		t.Error(err)
	}

	chunkedResourceSKUsClient, err := newSuccessfulFakeResourceSKUsClient(chunk(dataWrapper.Value, 10))
	if err != nil {
		t.Error(err)
	}

	ctx := context.Background()

	cases := map[string]struct {
		newCacheFunc NewCacheFunc
	}{
		"resourceSKUsClient": {
			newCacheFunc: func(_ context.Context, _ ...Option) (*Cache, error) {
				return NewCache(ctx, WithResourceSKUsClient(resourceSKUsClient), WithLocation("eastus"))
			},
		},
		"chunkedResourceSKUsClient": {
			newCacheFunc: func(_ context.Context, _ ...Option) (*Cache, error) {
				return NewCache(ctx, WithResourceSKUsClient(chunkedResourceSKUsClient), WithLocation("eastus"))
			},
		},
		"wrappedClient": {
			newCacheFunc: func(_ context.Context, _ ...Option) (*Cache, error) {
				return NewCache(ctx, WithClient(fakeClient), WithLocation("eastus"))
			},
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			cache, err := tc.newCacheFunc(ctx)
			if err != nil {
				t.Error(err)
			}
			t.Run("virtual machines", func(t *testing.T) {
				t.Run("expect 4 virtual machine skus", func(t *testing.T) {
					if len(cache.GetVirtualMachines(ctx)) != expectedVirtualMachinesCount {
						t.Errorf("expected %d virtual machine skus but found %d", expectedVirtualMachinesCount, len(cache.GetVirtualMachines(ctx)))
					}
				})

				t.Run("standard_d4s_v3", func(t *testing.T) {
					errCapabilityValueNil := &ErrCapabilityValueParse{}
					errCapabilityNotFound := &ErrCapabilityNotFound{}

					sku, err := cache.Get(ctx, "standard_d4s_v3", VirtualMachines, "eastus")
					if err != nil {
						t.Errorf("expected to find virtual machine sku standard_d4s_v3")
					}
					if name := sku.GetName(); !strings.EqualFold(name, "standard_d4s_v3") {
						t.Errorf("expected standard_d4s_v3 to have name standard_d4s_v3, got: '%s'", name)
					}
					if skuFamily := sku.GetFamilyName(); !strings.EqualFold(skuFamily, "standardDSv3Family") {
						t.Errorf("expected standard_d4s_v3 to have name standardDSv3Family, got: '%s'", skuFamily)
					}
					if skuSize := sku.GetSize(); !strings.EqualFold(skuSize, "d4s_v3") {
						t.Errorf("expected standard_d4s_v3 to have name d4s_v3 size, got: '%s'", skuSize)
					}
					if resourceType := sku.GetResourceType(); resourceType != VirtualMachines {
						t.Errorf("expected standard_d4s_v3 to have resourceType virtual machine, got: '%s'", resourceType)
					}
					if cpu, err := sku.VCPU(); cpu != 4 || err != nil {
						t.Errorf("expected standard_d4s_v3 to have 4 vCPUs and parse successfully, got value '%d' and error '%s'", cpu, err)
					}
					if memory, err := sku.Memory(); memory != 16 || err != nil {
						t.Errorf("expected standard_d4s_v3 to have 16GB of memory and parse successfully, got value '%f' and error '%s'", memory, err)
					}
					if quantity, err := sku.GetCapabilityIntegerQuantity("ShouldNotBePresent"); quantity != -1 || !errors.As(err, &errCapabilityNotFound) {
						t.Errorf("expected standard_d4s_v3 not to have a non-existent capability, got value '%d' and error '%s'", quantity, err)
					}
					if quantity, err := sku.GetCapabilityIntegerQuantity("PremiumIO"); quantity != -1 || !errors.As(err, &errCapabilityValueNil) {
						t.Errorf("expected standard_d4s_v3 to fail parsing value for boolean premiumIO as int, got value '%d' and error '%s'", quantity, err)
					}
					if !sku.HasZonalCapability(UltraSSDAvailable) {
						t.Errorf("expected standard_d4s_v3 to support ultra ssd")
					}
					if sku.HasZonalCapability("NotExistingCapability") {
						t.Errorf("expected standard_d4s_v3 not to support non-existent capability")
					}
					if !sku.HasCapability(EphemeralOSDisk) {
						t.Errorf("expected standard_d4s_v3 to support ephemeral os")
					}
					if !sku.IsAcceleratedNetworkingSupported() {
						t.Errorf("expected standard_d4s_v3 to support accelerated networking")
					}
					if cpuArch, err := sku.GetCPUArchitectureType(); err != nil || cpuArch != x64ArchType {
						t.Errorf("expected standard_d4s_v3 to have x64 cpuArchitectureType")
					}
					if !sku.IsPremiumIO() {
						t.Errorf("expected standard_d4s_v3 to support PremiumIO")
					}
					if !sku.IsHyperVGen1Supported() {
						t.Errorf("expected standard_d4s_v3 to support hyper v gen1")
					}
					if !sku.IsHyperVGen2Supported() {
						t.Errorf("expected standard_d4s_v3 to support hyper v gen2")
					}
					if !sku.HasCapability(EncryptionAtHost) {
						t.Errorf("expected standard_d4s_v3 to support encryption at host")
					}
					if !sku.IsAvailable("eastus") {
						t.Errorf("expected standard_d4s_v3 to be available in eastus")
					}
					if sku.IsRestricted("eastus") {
						t.Errorf("expected standard_d4s_v3 to be unrestricted in eastus")
					}
					if sku.IsAvailable("westus2") {
						t.Errorf("expected standard_d4s_v3 not to be available in westus2")
					}
					if sku.IsRestricted("westus2") {
						t.Errorf("expected standard_d4s_v3 not to be restricted in westus2")
					}
					if quantity, err := sku.MaxResourceVolumeMB(); quantity != 32768 || errors.As(err, &errCapabilityNotFound) {
						t.Errorf("expected standard_d4s_v3 to have 32768 MB of temporary disk, got value '%d' and error '%s'", quantity, err)
					}
					if isSupported, err := sku.HasCapabilityWithMinCapacity("MaxResourceVolumeMB", 32768); !isSupported || err != nil {
						t.Errorf("expected standard_d4s_v3 to  fit 32GB temp disk, got '%t', error: %s", isSupported, err)
					}
					if isSupported, err := sku.HasCapabilityWithMinCapacity("MaxResourceVolumeMB", 32769); isSupported || err != nil {
						t.Errorf("expected standard_d4s_v3 not to fit 32GB  +1 byte temp disk, got '%t', error: %s", isSupported, err)
					}
					hasV1 := !sku.HasCapabilityWithSeparator(HyperVGenerations, "V1")
					hasV2 := !sku.HasCapabilityWithSeparator(HyperVGenerations, "V2")
					if hasV1 || hasV2 {
						t.Errorf("expected standard_d4s_v3 to support hyper-v generation v1 and v2, got v1: '%t' , v2: '%t'", hasV1, hasV2)
					}
				})

				t.Run("standard_d2_v2", func(t *testing.T) {
					errCapabilityValueNil := &ErrCapabilityValueParse{}
					errCapabilityNotFound := &ErrCapabilityNotFound{}

					sku, err := cache.Get(ctx, "Standard_D2_v2", VirtualMachines, "eastus")
					if err != nil {
						t.Errorf("expected to find virtual machine sku standard_d2_v2")
					}
					if name := sku.GetName(); !strings.EqualFold(name, "standard_d2_v2") {
						t.Errorf("expected standard_d2_v2 to have name standard_d2_v2, got: '%s'", name)
					}
					if skuFamily := sku.GetFamilyName(); !strings.EqualFold(skuFamily, "standardDv2Family") {
						t.Errorf("expected standard_d2_v2 to have name standardDv2Family, got: '%s'", skuFamily)
					}
					if skuSize := sku.GetSize(); !strings.EqualFold(skuSize, "d2_v2") {
						t.Errorf("expected standard_d2_v2 to have name d2_v2 size, got: '%s'", skuSize)
					}
					if resourceType := sku.GetResourceType(); resourceType != VirtualMachines {
						t.Errorf("expected standard_d2_v2 to have resourceType virtual machine, got: '%s'", resourceType)
					}
					if cpu, err := sku.VCPU(); cpu != 2 || err != nil {
						t.Errorf("expected standard_d2_v2 to have 2 vCPUs and parse successfully, got value '%d' and error '%s'", cpu, err)
					}
					if memory, err := sku.Memory(); memory != 7 || err != nil {
						t.Errorf("expected standard_d2_v2 to have 7GB of memory and parse successfully, got value '%f' and error '%s'", memory, err)
					}
					if quantity, err := sku.GetCapabilityIntegerQuantity("ShouldNotBePresent"); quantity != -1 ||
						!errors.As(err, &errCapabilityNotFound) ||
						err.Error() != shouldNotBePresentCapabilityNotFoundErr {
						t.Errorf("expected standard_d2_v2 not to have a non-existent capability, got value '%d' and error '%s'", quantity, err)
					}
					if quantity, err := sku.GetCapabilityIntegerQuantity(CapabilityPremiumIO); quantity != -1 ||
						!errors.As(err, &errCapabilityValueNil) ||
						err.Error() != premiumIOCapabilityValueParseErr {
						t.Errorf("expected standard_d2_v2 to fail parsing value for boolean premiumIO as int, got value '%d' and error '%s'", quantity, err)
					}
					if sku.HasZonalCapability(UltraSSDAvailable) {
						t.Errorf("expected standard_d2_v2 not to support ultra ssd")
					}
					if sku.HasZonalCapability("NotExistingCapability") {
						t.Errorf("expected standard_d2_v2 not to support non-existent capability")
					}
					if sku.HasCapability(EphemeralOSDisk) {
						t.Errorf("expected standard_d2_v2 not to support ephemeral os")
					}
					if !sku.IsAcceleratedNetworkingSupported() {
						t.Errorf("expected standard_d2_v2 to support accelerated networking")
					}
					if cpuArch, err := sku.GetCPUArchitectureType(); err != nil || cpuArch != x64ArchType {
						t.Errorf("expected standard_d2_v2 to have x64 cpuArchitectureType")
					}
					if sku.IsPremiumIO() {
						t.Errorf("expected standard_d2_v2 to not support PremiumIO")
					}
					if !sku.IsHyperVGen1Supported() {
						t.Errorf("expected standard_d2_v2 to support hyper v gen1")
					}
					if sku.IsHyperVGen2Supported() {
						t.Errorf("expected standard_d2_v2 not to support hyper v gen2")
					}
					if sku.HasCapability(EncryptionAtHost) {
						t.Errorf("expected standard_d2_v2 not to support encryption at host")
					}
					if !sku.IsAvailable("eastus") {
						t.Errorf("expected standard_d2_v2 to be available in eastus")
					}
					if sku.IsRestricted("eastus") {
						t.Errorf("expected standard_d2_v2 to be unrestricted in eastus")
					}
					if sku.IsAvailable("westus2") {
						t.Errorf("expected standard_d2_v2 not to be available in westus2")
					}
					if sku.IsRestricted("westus2") {
						t.Errorf("expected standard_d2_v2 not to be restricted in westus2")
					}
					if quantity, err := sku.MaxResourceVolumeMB(); quantity != 102400 || errors.As(err, &errCapabilityNotFound) {
						t.Errorf("expected standard_d2_v2 to have 102400 MB of temporary disk, got value '%d' and error '%s'", quantity, err)
					}
					if isSupported, err := sku.HasCapabilityWithMinCapacity("MemoryGB", 1000); isSupported || err != nil {
						t.Errorf("expected standard_d2_v2 not to have 1000GB of memory, got '%t', error: %s", isSupported, err)
					}
					hasV1 := !sku.HasCapabilityWithSeparator(HyperVGenerations, "V1")
					hasV2 := sku.HasCapabilityWithSeparator(HyperVGenerations, "V2")
					if hasV1 || hasV2 {
						t.Errorf("expected standard_d2_v2 to support hyper-v generation v1 but not v2, got v1: '%t' , v2: '%t'", hasV1, hasV2)
					}
				})

				t.Run("Standard_NV6", func(t *testing.T) {
					errCapabilityValueNil := &ErrCapabilityValueParse{}
					errCapabilityNotFound := &ErrCapabilityNotFound{}

					sku, err := cache.Get(ctx, "Standard_NV6", VirtualMachines, "eastus")
					if err != nil {
						t.Errorf("expected to find virtual machine sku Standard_NV6")
					}
					if name := sku.GetName(); !strings.EqualFold(name, "standard_nv6") {
						t.Errorf("expected standard_nv6 to have name standard_nv6, got: '%s'", name)
					}
					if skuFamily := sku.GetFamilyName(); !strings.EqualFold(skuFamily, "standardNVFamily") {
						t.Errorf("expected standard_nv6 to have name standardNVFamily, got: '%s'", skuFamily)
					}
					if skuSize := sku.GetSize(); !strings.EqualFold(skuSize, "nv6") {
						t.Errorf("expected standard_nv6 to have name nv6 size, got: '%s'", skuSize)
					}
					if resourceType := sku.GetResourceType(); resourceType != VirtualMachines {
						t.Errorf("expected standard_nv6 to have resourceType virtual machine, got: '%s'", resourceType)
					}
					if cpu, err := sku.VCPU(); cpu != 6 || err != nil {
						t.Errorf("expected standard_nv6 to have 6 vCPUs and parse successfully, got value '%d' and error '%s'", cpu, err)
					}
					if cpu, err := sku.GPU(); cpu != 1 || err != nil {
						t.Errorf("expected standard_nv6 to have 1 GPUs and parse successfully, got value '%d' and error '%s'", cpu, err)
					}
					if memory, err := sku.Memory(); memory != 56 || err != nil {
						t.Errorf("expected standard_nv6 to have 56GB of memory and parse successfully, got value '%f' and error '%s'", memory, err)
					}
					if quantity, err := sku.GetCapabilityIntegerQuantity("ShouldNotBePresent"); quantity != -1 ||
						!errors.As(err, &errCapabilityNotFound) ||
						err.Error() != shouldNotBePresentCapabilityNotFoundErr {
						t.Errorf("expected standard_nv6 not to have a non-existent capability, got value '%d' and error '%s'", quantity, err)
					}
					if quantity, err := sku.GetCapabilityIntegerQuantity(CapabilityPremiumIO); quantity != -1 ||
						!errors.As(err, &errCapabilityValueNil) ||
						err.Error() != premiumIOCapabilityValueParseErr {
						t.Errorf("expected standard_nv6 to fail parsing value for boolean premiumIO as int, got value '%d' and error '%s'", quantity, err)
					}
					if sku.HasZonalCapability(UltraSSDAvailable) {
						t.Errorf("expected standard_nv6 not to support ultra ssd")
					}
					if sku.HasZonalCapability("NotExistingCapability") {
						t.Errorf("expected standard_nv6 not to support non-existent capability")
					}
					if sku.HasCapability(EphemeralOSDisk) {
						t.Errorf("expected standard_nv6 not to support ephemeral os")
					}
					if sku.IsAcceleratedNetworkingSupported() {
						t.Errorf("expected standard_nv6 to not support accelerated networking")
					}
					if cpuArch, err := sku.GetCPUArchitectureType(); err != nil || cpuArch != x64ArchType {
						t.Errorf("expected standard_nv6 to have x64 cpuArchitectureType")
					}
					if sku.IsPremiumIO() {
						t.Errorf("expected standard_nv6 to not support PremiumIO")
					}
					if !sku.IsHyperVGen1Supported() {
						t.Errorf("expected standard_nv6 to support hyper v gen1")
					}
					if sku.IsHyperVGen2Supported() {
						t.Errorf("expected standard_nv6 not to support hyper v gen2")
					}
					if sku.HasCapability(EncryptionAtHost) {
						t.Errorf("expected standard_nv6 not to support encryption at host")
					}
					if !sku.IsAvailable("eastus") {
						t.Errorf("expected standard_nv6 to be available in eastus")
					}
					if sku.IsRestricted("eastus") {
						t.Errorf("expected standard_nv6 to be unrestricted in eastus")
					}
					if sku.IsAvailable("westus2") {
						t.Errorf("expected standard_nv6 not to be available in westus2")
					}
					if sku.IsRestricted("westus2") {
						t.Errorf("expected standard_nv6 not to be restricted in westus2")
					}
					if quantity, err := sku.MaxResourceVolumeMB(); quantity != 389120 || errors.As(err, &errCapabilityNotFound) {
						t.Errorf("expected standard_nv6 to have 389120 MB of temporary disk, got value '%d' and error '%s'", quantity, err)
					}
					if isSupported, err := sku.HasCapabilityWithMinCapacity("MemoryGB", 1000); isSupported || err != nil {
						t.Errorf("expected standard_nv6 not to have 1000GB of memory, got '%t', error: %s", isSupported, err)
					}
					hasV1 := !sku.HasCapabilityWithSeparator(HyperVGenerations, "V1")
					hasV2 := sku.HasCapabilityWithSeparator(HyperVGenerations, "V2")
					if hasV1 || hasV2 {
						t.Errorf("expected standard_nv6 to support hyper-v generation v1 but not v2, got v1: '%t' , v2: '%t'", hasV1, hasV2)
					}
				})

				t.Run("standard_D13_v2_promo", func(t *testing.T) {
					errCapabilityNotFound := &ErrCapabilityNotFound{}
					sku, err := cache.Get(ctx, "standard_D13_v2_promo", VirtualMachines, "eastus")
					if err != nil {
						t.Errorf("expected to find virtual machine sku standard_D13_v2_promo")
					}
					if sku.IsAvailable("eastus") {
						t.Errorf("expected standard_D13_v2_promo to be unavailable in eastus")
					}
					if !sku.IsRestricted("eastus") {
						t.Errorf("expected standard_D13_v2_promo to be restricted in eastus")
					}
					if sku.IsAvailable("westus2") {
						t.Errorf("expected standard_D13_v2_promo not to be available in westus2")
					}
					if sku.IsRestricted("westus2") {
						t.Errorf("expected standard_D13_v2_promo not to be restricted in westus2")
					}
					if cpuArch, err := sku.GetCPUArchitectureType(); !errors.As(err, &errCapabilityNotFound) || cpuArch != "" {
						t.Errorf("expected standard_D13_v2_promo to not have cpuArchitectureType, got %s as cpuArchType with error as %s", cpuArch, err)
					}
				})
			})

			t.Run("availability zones", func(t *testing.T) {
				if diff := cmp.Diff(cache.GetAvailabilityZones(ctx), expectedAvailabilityZones, []cmp.Option{
					cmpopts.EquateEmpty(),
					cmpopts.SortSlices(func(a, b string) bool {
						return a < b
					}),
				}...); diff != "" {
					t.Errorf("expected and actual availability zones mismatch: %s", diff)
				}
			})
		})
	}
}
