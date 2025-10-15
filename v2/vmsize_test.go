package skewer

import (
	"fmt"
	"testing"

	"github.com/Azure/skewer/v2/testdata"
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
