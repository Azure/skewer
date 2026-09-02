package skewer

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// This file adds support for more capabilities based on VM naming conventions that includes vmsize parsing.
// VM naming conventions are documented at: https://docs.microsoft.com/en-us/azure/virtual-machines/vm-naming-conventions
// Note: Some common capabilities like familyName and VCPUs, which can also be
// fetched using the ResourceSKU API, are not included here. They can be found in sku.go.

var skuSizeScheme = regexp.MustCompile(
	`^([A-Z])([A-Z]?)([A-Z]?)([0-9]+)-?((?:[0-9]+)?)((?:[abcdeiflmnotspPr]+|C+|NP)?)_?(?:NDR_)?(?:((?:xl_)?[A-Z]+[0-9]+[A-Z]*)_?)?(_cc_)?(_[0-9]+_)?(_MI300X_)?(_H100_)?((?:[vV][1-9])?)?(_Promo)?$`,
)

// Azure only reports nested virtualization support for new SKUs.
// Keep this name-based check as a fallback when the capability is absent.
var nestedVirtualizationEnabledSKUs = []*regexp.Regexp{
	regexp.MustCompile(`^standard_d\d+s?_v3$`),                            // d<digits>[s] https://learn.microsoft.com/en-us/azure/virtual-machines/dv3-dsv3-series
	regexp.MustCompile(`^standard_d\d+s?_v4$`),                            // d<digits>[s] https://learn.microsoft.com/en-us/azure/virtual-machines/dv4-dsv4-series
	regexp.MustCompile(`^standard_d\d+s?_v5$`),                            // d<digits>[s] https://learn.microsoft.com/en-us/azure/virtual-machines/dv5-dsv5-series
	regexp.MustCompile(`^standard_d\d+ds?_v4$`),                           // d<digits>d[s] https://learn.microsoft.com/en-us/azure/virtual-machines/ddv4-ddsv4-series
	regexp.MustCompile(`^standard_d\d+ds?_v5$`),                           // d<digits>d[s] https://learn.microsoft.com/en-us/azure/virtual-machines/ddv5-ddsv5-series
	regexp.MustCompile(`^standard_d\d+ad?s_v5$`),                          // d<digits>a[d]s https://learn.microsoft.com/en-us/azure/virtual-machines/dasv5-dadsv5-series
	regexp.MustCompile(`^standard_d\d+l?ds?_v5$`),                         // d<digits>d[s] https://learn.microsoft.com/en-us/azure/virtual-machines/ddv5-ddsv5-series
	regexp.MustCompile(`^standard_d\d+l?s_v5$`),                           // d<digits>l https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/general-purpose/dlsv5-series
	regexp.MustCompile(`^standard_d\d+(s|ds|ls|lds|as|ads|als|alds)_v6$`), // New v6 patterns (precise to avoid 'p' variants)
	regexp.MustCompile(`^standard_dc\d+ad?s_cc_v5$`),                      // dc<digits>a[d]s https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/general-purpose/dcasccv5-series
	regexp.MustCompile(`^standard_d\d+ads_v7$`),                           // d<digits>ads https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/general-purpose/dadsv7-series
	regexp.MustCompile(`^standard_d\d+as_v7$`),                            // d<digits>as https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/general-purpose/dasv7-series
	regexp.MustCompile(`^standard_d\d+als_v7$`),                           // d<digits>als https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/general-purpose/dalsv7-series
	regexp.MustCompile(`^standard_d\d+alds_v7$`),                          // d<digits>alds https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/general-purpose/daldsv7-series

	regexp.MustCompile(`^standard_f\d+s_v2$`),       // f<digits>s https://learn.microsoft.com/en-us/azure/virtual-machines/fsv2-series
	regexp.MustCompile(`^standard_f\d+a[lm]?s_v6$`), // f<digits>a[l,m]s_v6 https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/compute-optimized/falsv6-series
	regexp.MustCompile(`^standard_f\d+as_v7$`),      // f<digits>as https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/compute-optimized/fasv7-series
	regexp.MustCompile(`^standard_f\d+als_v7$`),     // f<digits>als https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/compute-optimized/falsv7-series
	regexp.MustCompile(`^standard_f\d+ams_v7$`),     // f<digits>ams https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/compute-optimized/famsv7-series
	regexp.MustCompile(`^standard_f\d+ads_v7$`),     // f<digits>ads https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/compute-optimized/fadsv7-series
	regexp.MustCompile(`^standard_f\d+alds_v7$`),    // f<digits>alds https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/compute-optimized/faldsv7-series
	regexp.MustCompile(`^standard_f\d+amds_v7$`),    // f<digits>amds https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/compute-optimized/famdsv7-series
	regexp.MustCompile(`^standard_fx\d+mds$`),       // fx<digits>mds https://learn.microsoft.com/en-us/azure/virtual-machines/fx-series
	regexp.MustCompile(`^standard_fx\d+md?s_v2$`),   // fx<digits>m[d]s_v2 https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/compute-optimized/fx-family

	regexp.MustCompile(`^standard_e\d+i?s?_v3$`),      // e<digits>[i][s] https://learn.microsoft.com/en-us/azure/virtual-machines/ev3-esv3-series
	regexp.MustCompile(`^standard_e\d+i?s?_v4$`),      // e<digits>[i][s] https://learn.microsoft.com/en-us/azure/virtual-machines/ev4-esv4-series
	regexp.MustCompile(`^standard_e\d+i?s?_v5$`),      // e<digits>[i][s] https://learn.microsoft.com/en-us/azure/virtual-machines/ev5-esv5-series
	regexp.MustCompile(`^standard_e\d+(i|a)?d?s_v6$`), // e<digits>[i,a][d]s_v6 https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/memory-optimized/esv6-series
	regexp.MustCompile(`^standard_e\d+i?ds?_v4$`),     // e<digits>[i]d[s] https://learn.microsoft.com/en-us/azure/virtual-machines/edv4-edsv4-series
	regexp.MustCompile(`^standard_e\d+i?ds?_v5$`),     // e<digits>[i]d[s] https://learn.microsoft.com/en-us/azure/virtual-machines/edv5-edsv5-series
	regexp.MustCompile(`^standard_e\d+i?ad?s_v5$`),    // e<digits>[i]a[d]s https://learn.microsoft.com/en-us/azure/virtual-machines/easv5-eadsv5-series
	regexp.MustCompile(`^standard_e\d+bd?s_v5$`),      // e<digits>b[d]s https://learn.microsoft.com/en-us/azure/virtual-machines/ebdsv5-ebsv5-series
	regexp.MustCompile(`^standard_e\d+bs_v6$`),        // e<digits>bs https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/memory-optimized/ebsv6-series
	regexp.MustCompile(`^standard_e\d+bds_v6$`),       // e<digits>bds https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/memory-optimized/ebdsv6-series
	regexp.MustCompile(`^standard_e\d+as_v7$`),        // e<digits>as https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/memory-optimized/easv7-series
	regexp.MustCompile(`^standard_e\d+ads_v7$`),       // e<digits>ads https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/memory-optimized/eadsv7-series
	regexp.MustCompile(`^standard_ec\d+ad?s_cc_v5$`),  // ec<digits>a[d]s_cc_v5 https://learn.microsoft.com/en-us/azure/virtual-machines/ecasccv5-ecadsccv5-series

	regexp.MustCompile(`^standard_m\d+(m?s?|ls|ts)$`), // m<digits>[m,l,t][s] https://learn.microsoft.com/en-us/azure/virtual-machines/m-series

	regexp.MustCompile(`^standard_l\d+s_v3$`),   // l<digits>s https://learn.microsoft.com/en-us/azure/virtual-machines/lsv3-series
	regexp.MustCompile(`^standard_l\d+as_v3$`),  // l<digits>as https://learn.microsoft.com/en-us/azure/virtual-machines/lasv3-series
	regexp.MustCompile(`^standard_l\d+s_v4$`),   // l<digits>s https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/storage-optimized/lsv4-series
	regexp.MustCompile(`^standard_l\d+as_v4$`),  // l<digits>as https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/storage-optimized/lasv4-series
	regexp.MustCompile(`^standard_l\d+aos_v4$`), // l<digits>aos https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/storage-optimized/laosv4-series

	regexp.MustCompile(`^standard_nv\d+ads_v710_v5$`), // nv<digits>ads_V710_v5 https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/gpu-accelerated/nvadsv710-v5-series
}

// unParsableVMSizes map holds vmSize strings that cannot be easily parsed with skuSizeScheme.
var unParsableVMSizes = map[string]VMSizeType{
	"M416s_8_v2": {
		Family:                      "M",
		Subfamily:                   nil,
		Cpus:                        "416",
		CpusConstrained:             nil,
		AdditiveFeatures:            []rune{'s'},
		AcceleratorType:             nil,
		ConfidentialChildCapability: false,
		Version:                     "v2",
		PromoVersion:                false,
		Series:                      "Ms_v2",
	},
}

// acceleratorTypeBySize maps sizes whose name carries no accelerator to their GPU.
// AcceleratorType is parsed out of the size name, so these would be nil otherwise.
// The SKUs API only gives a GPU count, no model. Some family strings do carry the
// chip, but not the ones listed here, and the format varies too much to parse.
// Values come from the per-series pages on MS Learn.
var acceleratorTypeBySize = map[string]string{
	// https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/gpu-accelerated/ncv3-series
	"NC6s_v3":   "V100",
	"NC12s_v3":  "V100",
	"NC24s_v3":  "V100",
	"NC24rs_v3": "V100",

	// https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/gpu-accelerated/ndv2-series
	"ND40rs_v2": "V100",

	// https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/gpu-accelerated/ndasra100v4-series
	"ND96asr_v4": "A100",

	// https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/gpu-accelerated/nvv3-series
	"NV12s_v3": "M60",
	"NV24s_v3": "M60",
	"NV48s_v3": "M60",

	// https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/gpu-accelerated/nvv4-series
	"NV4as_v4":  "MI25",
	"NV8as_v4":  "MI25",
	"NV16as_v4": "MI25",
	"NV32as_v4": "MI25",
}

type VMSizeType struct {
	Family                      string
	Subfamily                   *string
	Cpus                        string
	CpusConstrained             *string
	AdditiveFeatures            []rune
	AcceleratorType             *string
	ConfidentialChildCapability bool
	Version                     string
	PromoVersion                bool
	MI300Series                 bool
	H100Series                  bool
	Series                      string
}

// parseVMSize parses the VM size and returns the parts as a map.
func parseVMSize(vmSizeName string) ([]string, error) {
	parts := skuSizeScheme.FindStringSubmatch(vmSizeName)
	if parts == nil || len(parts) < 10 {
		return nil, fmt.Errorf("could not parse VM size %s", vmSizeName)
	}
	return parts, nil
}

// GetVMSize is a helper function used by GetVMSize() in sku.go
func GetVMSize(vmSizeName string) (*VMSizeType, error) {
	vmSize := VMSizeType{}

	parts, err := parseVMSize(vmSizeName)
	if err != nil {
		if vmSizeVal, ok := unParsableVMSizes[vmSizeName]; ok {
			return &vmSizeVal, nil
		}
		return nil, err
	}

	// [Family] - ([A-Z]): Captures a single uppercase letter.
	vmSize.Family = parts[1]

	// [Sub-family]* - ([A-Z]?): Optionally captures another uppercase letter.
	if len(parts[2]) > 0 {
		var subfamilyStr string
		if len(parts[3]) > 0 {
			subfamilyStr = parts[2] + parts[3]
		} else {
			subfamilyStr = parts[2]
		}
		vmSize.Subfamily = &subfamilyStr
	}

	// [# of vCPUs] - ([0-9]+): Captures one or more digits.
	vmSize.Cpus = parts[4]

	// [Constrained vCPUs]*
	// -?: Optionally captures a hyphen.
	// ((?:[0-9]+)?): Optionally captures another sequence of one or more digits.
	if len(parts[5]) > 0 {
		_, err := strconv.Atoi(parts[5])
		if err != nil {
			return nil, fmt.Errorf("converting constrained CPUs, %w", err)
		}
		vmSize.CpusConstrained = &parts[5]
	}

	// [Additive Features]
	// ((?:[abcdilmtspPr]+|C+|NP)?): Captures a sequence of letters representing certain attributes.
	// It can capture combinations like 'abcdilmtspPr' or 'C+' or 'NP'.
	vmSize.AdditiveFeatures = []rune(parts[6])

	// [Accelerator Type]*
	// _?: Optionally captures an underscore.
	// (?:((?:xl_)?[A-Z]+[0-9]+[A-Z]*)_?)?: Optionally captures the accelerator type.
	// Covers the common "<letters><digits>" form (e.g. "A100", "T4"), names with
	// trailing letters (e.g. "MI300X"), and the optional "xl_" size descriptor
	// used by some GPU SKUs (e.g. "xl_RTXPRO6000BSE" in NC288ds_xl_RTXPRO6000BSE_v6).
	// A trailing optional underscore is also captured.
	if len(parts[7]) > 0 {
		// Strip the optional size descriptor (e.g. "xl_") so AcceleratorType reflects
		// the accelerator name only.
		accelerator := strings.TrimPrefix(parts[7], "xl_")
		vmSize.AcceleratorType = &accelerator
	} else if accelerator, ok := acceleratorTypeBySize[strings.TrimSuffix(vmSizeName, "_Promo")]; ok {
		vmSize.AcceleratorType = &accelerator
	}

	// [Confidential Child Capability]* - only AKS
	// (_cc_)?: Optionally captures the string "cc" with underscores on both sides.
	if parts[8] == "_cc_" {
		vmSize.ConfidentialChildCapability = true
	}

	// parts slice at index 8 disambiguates more enhanced memory and I/O capabilities
	// for Standard M memory-optimized VM series.
	// For example:
	// 1 in Standard_M96s_1_v3
	// and 2 in Standard_M96s_2_v3
	// Ref: https://learn.microsoft.com/en-us/azure/virtual-machines/msv3-mdsv3-medium-series

	// [MI300X]*
	// (_MI300X_)?: Optionally captures the string "_MI300X".
	// This is used to identify the MI300 series of VMs.
	if parts[10] == "MI300X" {
		vmSize.MI300Series = true
	}

	// [H100]*
	// (_H100_)?: Optionally captures the string "_H100".
	// This is used to identify the H100 series of VMs.
	if parts[11] == "H100" {
		vmSize.H100Series = true
	}

	// [Version]*
	// Optionally captures the pattern 'v' or 'V' followed by a digit from 1 to 9.
	vmSize.Version = parts[12]

	// [Promo]*
	// (_Promo)?: Optionally captures the string "_Promo".
	if parts[13] == "_Promo" {
		vmSize.PromoVersion = true
	}

	// [Series]
	subfamily := ""
	if vmSize.Subfamily != nil {
		subfamily = *vmSize.Subfamily
	}
	version := ""
	if len(vmSize.Version) > 0 {
		version = "_" + vmSize.Version
	}
	vmSize.Series = vmSize.Family + subfamily + string(vmSize.AdditiveFeatures) + version

	return &vmSize, nil
}

func supportsNestedVirtualization(vmSizeName string) bool {
	standardizedVMSizeName := "standard_" + strings.ToLower(vmSizeName)
	for _, pattern := range nestedVirtualizationEnabledSKUs {
		if pattern.MatchString(standardizedVMSizeName) {
			return true
		}
	}
	return false
}
