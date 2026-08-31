package skewer

import (
	"fmt"
	"testing"

	"github.com/Azure/skewer/testdata"
	"github.com/stretchr/testify/assert"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
)

// TestParseVMSize tests the parseVMSize function. It uses testdata from generated_vmsize_testdata.go
// This test validates the parsing capability and not the actual values.
func TestParseVMSize(t *testing.T) {
	total := len(testdata.SKUData)
	fail := 0
	for skuName, tc := range testdata.SKUData {
		if _, err := parseVMSize(tc.Size); err != nil {
			if _, ok := unParsableVMSizes[tc.Size]; !ok {
				t.Errorf("parsing fails for for sku %s, err: %v", skuName, err)
				fail += 1
			}
		}
	}
	t.Logf("Passed SKUs: %d, Failed SKUs: %d", total-fail, fail)
}

// Define the test cases for get() methods in vmsize.go
var testCases = []struct {
	name       string
	size       string
	expectedVM *VMSizeType
	err        error
}{
	{
		name: "Standard_NV16as_v4",
		size: "NV16as_v4",
		expectedVM: &VMSizeType{
			Family:                      "N",
			Subfamily:                   to.Ptr("V"),
			Cpus:                        "16",
			CpusConstrained:             nil,
			AdditiveFeatures:            []rune{'a', 's'},
			AcceleratorType:             nil,
			ConfidentialChildCapability: false,
			Version:                     "v4",
			PromoVersion:                false,
			Series:                      "NVas_v4",
		},
		err: nil,
	},
	{
		name: "Standard_M16ms",
		size: "M16ms_v2",
		expectedVM: &VMSizeType{
			Family:                      "M",
			Subfamily:                   nil,
			Cpus:                        "16",
			CpusConstrained:             nil,
			AdditiveFeatures:            []rune{'m', 's'},
			AcceleratorType:             nil,
			ConfidentialChildCapability: false,
			Version:                     "v2",
			PromoVersion:                false,
			Series:                      "Mms_v2",
		},
		err: nil,
	},
	{
		name: "Standard_NC4as_T4_v3",
		size: "NC4as_T4_v3",
		expectedVM: &VMSizeType{
			Family:                      "N",
			Subfamily:                   to.Ptr("C"),
			Cpus:                        "4",
			CpusConstrained:             nil,
			AdditiveFeatures:            []rune{'a', 's'},
			AcceleratorType:             to.Ptr("T4"),
			ConfidentialChildCapability: false,
			Version:                     "v3",
			PromoVersion:                false,
			Series:                      "NCas_v3",
		},
		err: nil,
	},
	{
		name: "Standard_M8-2ms",
		size: "M8-2ms_v2",
		expectedVM: &VMSizeType{
			Family:                      "M",
			Subfamily:                   nil,
			Cpus:                        "8",
			CpusConstrained:             to.Ptr("2"),
			AdditiveFeatures:            []rune{'m', 's'},
			AcceleratorType:             nil,
			ConfidentialChildCapability: false,
			Version:                     "v2",
			PromoVersion:                false,
			Series:                      "Mms_v2",
		},
		err: nil,
	},
	{
		name: "Standard_A4_v2",
		size: "A4_v2",
		expectedVM: &VMSizeType{
			Family:                      "A",
			Subfamily:                   nil,
			Cpus:                        "4",
			CpusConstrained:             nil,
			AdditiveFeatures:            []rune{},
			AcceleratorType:             nil,
			ConfidentialChildCapability: false,
			Version:                     "v2",
			PromoVersion:                false,
			Series:                      "A_v2",
		},
		err: nil,
	},
	{
		name: "Standard_EC48as_cc_v5",
		size: "EC48as_cc_v5",
		expectedVM: &VMSizeType{
			Family:                      "E",
			Subfamily:                   to.Ptr("C"),
			Cpus:                        "48",
			CpusConstrained:             nil,
			AdditiveFeatures:            []rune{'a', 's'},
			AcceleratorType:             nil,
			ConfidentialChildCapability: true,
			Version:                     "v5",
			PromoVersion:                false,
			Series:                      "ECas_v5",
		},
		err: nil,
	},
	{
		name: "Standard_NV24",
		size: "NV24",
		expectedVM: &VMSizeType{
			Family:                      "N",
			Subfamily:                   to.Ptr("V"),
			Cpus:                        "24",
			CpusConstrained:             nil,
			AdditiveFeatures:            []rune{},
			AcceleratorType:             nil,
			ConfidentialChildCapability: false,
			Version:                     "",
			PromoVersion:                false,
			Series:                      "NV",
		},
		err: nil,
	},
	{
		name: "Standard_D3_v2_Promo",
		size: "D3_v2_Promo",
		expectedVM: &VMSizeType{
			Family:                      "D",
			Subfamily:                   nil,
			Cpus:                        "3",
			CpusConstrained:             nil,
			AdditiveFeatures:            []rune{},
			AcceleratorType:             nil,
			ConfidentialChildCapability: false,
			Version:                     "v2",
			PromoVersion:                true,
			Series:                      "D_v2",
		},
		err: nil,
	},
	{
		name: "Standard_L12aos_v4",
		size: "L12aos_v4",
		expectedVM: &VMSizeType{
			Family:                      "L",
			Subfamily:                   nil,
			Cpus:                        "12",
			CpusConstrained:             nil,
			AdditiveFeatures:            []rune{'a', 'o', 's'},
			AcceleratorType:             nil,
			ConfidentialChildCapability: false,
			Version:                     "v4",
			PromoVersion:                false,
			Series:                      "Laos_v4",
		},
		err: nil,
	},
	{
		name: "Standard_D4fs_v6",
		size: "D4fs_v6",
		expectedVM: &VMSizeType{
			Family:                      "D",
			Subfamily:                   nil,
			Cpus:                        "4",
			CpusConstrained:             nil,
			AdditiveFeatures:            []rune{'f', 's'},
			AcceleratorType:             nil,
			ConfidentialChildCapability: false,
			Version:                     "v6",
			PromoVersion:                false,
			Series:                      "Dfs_v6",
		},
		err: nil,
	},
	{
		name: "Standard_D8ns_v6",
		size: "D8ns_v6",
		expectedVM: &VMSizeType{
			Family:                      "D",
			Subfamily:                   nil,
			Cpus:                        "8",
			CpusConstrained:             nil,
			AdditiveFeatures:            []rune{'n', 's'},
			AcceleratorType:             nil,
			ConfidentialChildCapability: false,
			Version:                     "v6",
			PromoVersion:                false,
			Series:                      "Dns_v6",
		},
		err: nil,
	},
	{
		name: "Standard_ND128isr_GB300_v6",
		size: "ND128isr_GB300_v6",
		expectedVM: &VMSizeType{
			Family:                      "N",
			Subfamily:                   to.Ptr("D"),
			Cpus:                        "128",
			CpusConstrained:             nil,
			AdditiveFeatures:            []rune{'i', 's', 'r'},
			AcceleratorType:             to.Ptr("GB300"),
			ConfidentialChildCapability: false,
			Version:                     "v6",
			PromoVersion:                false,
			Series:                      "NDisr_v6",
		},
		err: nil,
	},
	{
		name: "Standard_NC288ds_xl_RTXPRO6000BSE_v6",
		size: "NC288ds_xl_RTXPRO6000BSE_v6",
		expectedVM: &VMSizeType{
			Family:                      "N",
			Subfamily:                   to.Ptr("C"),
			Cpus:                        "288",
			CpusConstrained:             nil,
			AdditiveFeatures:            []rune{'d', 's'},
			AcceleratorType:             to.Ptr("RTXPRO6000BSE"),
			ConfidentialChildCapability: false,
			Version:                     "v6",
			PromoVersion:                false,
			Series:                      "NCds_v6",
		},
		err: nil,
	},
	{
		name: "Standard_NC24lds_xl_RTXPRO6000BSE_v6",
		size: "NC24lds_xl_RTXPRO6000BSE_v6",
		expectedVM: &VMSizeType{
			Family:                      "N",
			Subfamily:                   to.Ptr("C"),
			Cpus:                        "24",
			CpusConstrained:             nil,
			AdditiveFeatures:            []rune{'l', 'd', 's'},
			AcceleratorType:             to.Ptr("RTXPRO6000BSE"),
			ConfidentialChildCapability: false,
			Version:                     "v6",
			PromoVersion:                false,
			Series:                      "NClds_v6",
		},
		err: nil,
	},
	{
		name:       "Standard_inValid",
		size:       "inValid",
		expectedVM: nil,
		err:        fmt.Errorf("could not parse VM size inValid"),
	},
}

// Test_GetVMSize tests the GetVMSize() function.
func Test_GetVMSize(t *testing.T) {
	a := assert.New(t)
	for _, test := range testCases {
		vmSize, err := GetVMSize(test.size)
		a.Equal(test.err, err)
		if err != nil {
			continue
		}
		a.Equal(test.expectedVM.Family, vmSize.Family)
		a.Equal(test.expectedVM.Subfamily, vmSize.Subfamily)
		a.Equal(test.expectedVM.Cpus, vmSize.Cpus)
		a.Equal(test.expectedVM.CpusConstrained, vmSize.CpusConstrained)
		a.Equal(test.expectedVM.AdditiveFeatures, vmSize.AdditiveFeatures)
		a.Equal(test.expectedVM.AcceleratorType, vmSize.AcceleratorType)
		a.Equal(test.expectedVM.ConfidentialChildCapability, vmSize.ConfidentialChildCapability)
		a.Equal(test.expectedVM.Version, vmSize.Version)
		a.Equal(test.expectedVM.PromoVersion, vmSize.PromoVersion)
	}
}

func TestGetVMSizeNestedVirtualizationSupported(t *testing.T) {
	testCases := map[string]bool{
		"D2_v3":        true,
		"D64_v3":       true,
		"D4s_v3":       true,
		"D64d_v4":      true,
		"D64ds_v4":     true,
		"D8ds_v5":      true,
		"D96as_v5":     true,
		"D96ads_v5":    true,
		"D4als_v6":     true,
		"DC4as_cc_v5":  true,
		"DC2ads_cc_v5": true,
		"D4ads_v7":     true,
		"D16ads_v7":    true,
		"F4s_v2":       true,
		"F72s_v2":      true,
		"F8ams_v6":     true,
		"FX48mds":      true,
		"FX48ms_v2":    true,
		"FX4mds_v2":    true,
		"E64i_v3":      true,
		"E8is_v3":      true,
		"E64d_v4":      true,
		"E80ids_v4":    true,
		"E104is_v5":    true,
		"E112ias_v5":   true,
		"E8ads_v5":     true,
		"E64bds_v5":    true,
		"E8ads_v6":     true,
		"E8bs_v5":      true,
		"EC8ads_cc_v5": true,
		"M128m":        true,
		"M128ms":       true,
		"M64ls":        true,
		"L80s_v3":      true,
		"L8as_v3":      true,
		"DC48s_v3":     false,
		"E2a_v4":       false,
		"E2pds_v5":     false,
		"D4ps_v6":      false,
		"D2ds_v7":      false,
		"M128ns":       false,
		"D4as_v4":      false,
		"F4_v2":        false,
		"L8s_v2":       false,
		"NC4as_T4_v3":  false,
	}

	for vmSizeName, expected := range testCases {
		t.Run(vmSizeName, func(t *testing.T) {
			vmSize, err := GetVMSize(vmSizeName)
			assert.NoError(t, err)
			assert.Equal(t, expected, vmSize.NestedVirtualizationSupported)
		})
	}
}
