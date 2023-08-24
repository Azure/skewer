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
			family:                      "N",
			subfamily:                   to.Ptr("V"),
			cpus:                        "16",
			cpusConstrained:             nil,
			additiveFeatures:            []rune{'a', 's'},
			acceleratorType:             nil,
			confidentialChildCapability: false,
			version:                     "v4",
			promoVersion:                false,
			series:                      "NVas_v4",
		},
		err: nil,
	},
	{
		name: "Standard_M16ms",
		size: "M16ms_v2",
		expectedVM: &VMSizeType{
			family:                      "M",
			subfamily:                   nil,
			cpus:                        "16",
			cpusConstrained:             nil,
			additiveFeatures:            []rune{'m', 's'},
			acceleratorType:             nil,
			confidentialChildCapability: false,
			version:                     "v2",
			promoVersion:                false,
			series:                      "Mms_v2",
		},
		err: nil,
	},
	{
		name: "Standard_NC4as_T4_v3",
		size: "NC4as_T4_v3",
		expectedVM: &VMSizeType{
			family:                      "N",
			subfamily:                   to.Ptr("C"),
			cpus:                        "4",
			cpusConstrained:             nil,
			additiveFeatures:            []rune{'a', 's'},
			acceleratorType:             to.Ptr("T4"),
			confidentialChildCapability: false,
			version:                     "v3",
			promoVersion:                false,
			series:                      "NCas_v3",
		},
		err: nil,
	},
	{
		name: "Standard_M8-2ms",
		size: "M8-2ms_v2",
		expectedVM: &VMSizeType{
			family:                      "M",
			subfamily:                   nil,
			cpus:                        "8",
			cpusConstrained:             to.Ptr("2"),
			additiveFeatures:            []rune{'m', 's'},
			acceleratorType:             nil,
			confidentialChildCapability: false,
			version:                     "v2",
			promoVersion:                false,
			series:                      "Mms_v2",
		},
		err: nil,
	},
	{
		name: "Standard_A4_v2",
		size: "A4_v2",
		expectedVM: &VMSizeType{
			family:                      "A",
			subfamily:                   nil,
			cpus:                        "4",
			cpusConstrained:             nil,
			additiveFeatures:            []rune{},
			acceleratorType:             nil,
			confidentialChildCapability: false,
			version:                     "v2",
			promoVersion:                false,
			series:                      "A_v2",
		},
		err: nil,
	},
	{
		name: "Standard_EC48as_cc_v5",
		size: "EC48as_cc_v5",
		expectedVM: &VMSizeType{
			family:                      "E",
			subfamily:                   to.Ptr("C"),
			cpus:                        "48",
			cpusConstrained:             nil,
			additiveFeatures:            []rune{'a', 's'},
			acceleratorType:             nil,
			confidentialChildCapability: true,
			version:                     "v5",
			promoVersion:                false,
			series:                      "ECas_v5",
		},
		err: nil,
	},
	{
		name: "Standard_NV24",
		size: "NV24",
		expectedVM: &VMSizeType{
			family:                      "N",
			subfamily:                   to.Ptr("V"),
			cpus:                        "24",
			cpusConstrained:             nil,
			additiveFeatures:            []rune{},
			acceleratorType:             nil,
			confidentialChildCapability: false,
			version:                     "",
			promoVersion:                false,
			series:                      "NV",
		},
		err: nil,
	},
	{
		name: "Standard_D3_v2_Promo",
		size: "D3_v2_Promo",
		expectedVM: &VMSizeType{
			family:                      "D",
			subfamily:                   nil,
			cpus:                        "3",
			cpusConstrained:             nil,
			additiveFeatures:            []rune{},
			acceleratorType:             nil,
			confidentialChildCapability: false,
			version:                     "v2",
			promoVersion:                true,
			series:                      "D_v2",
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

// Test_getSize tests the getSize() function.
func Test_getVMsize(t *testing.T) {
	a := assert.New(t)
	for _, test := range testCases {
		vmSize, err := getVMSize(test.size)
		a.Equal(test.err, err)
		if err != nil {
			continue
		}
		a.Equal(test.expectedVM.family, vmSize.family)
		a.Equal(test.expectedVM.subfamily, vmSize.subfamily)
		a.Equal(test.expectedVM.cpus, vmSize.cpus)
		a.Equal(test.expectedVM.cpusConstrained, vmSize.cpusConstrained)
		a.Equal(test.expectedVM.additiveFeatures, vmSize.additiveFeatures)
		a.Equal(test.expectedVM.acceleratorType, vmSize.acceleratorType)
		a.Equal(test.expectedVM.confidentialChildCapability, vmSize.confidentialChildCapability)
		a.Equal(test.expectedVM.version, vmSize.version)
		a.Equal(test.expectedVM.promoVersion, vmSize.promoVersion)
	}
}
