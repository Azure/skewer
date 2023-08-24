package skewer

import (
	"fmt"
	"regexp"
	"strconv"
)

// This file adds support for more capabilities based on VM naming conventions that includes vmsize parsing.
// VM naming conventions are documented at: https://docs.microsoft.com/en-us/azure/virtual-machines/vm-naming-conventions
// Note: Some common capabilities like familyName and VCPUs, which can also be
// fetched using the ResourceSKU API, are not included here. They can be found in sku.go.

var skuSizeScheme = regexp.MustCompile(
	`^([A-Z])([A-Z]?)([0-9]+)-?((?:[0-9]+)?)((?:[abcdilmtspPr]+|C+|NP)?)_?(?:([A-Z][0-9]+)_?)?(_cc_)?((?:[vV][1-9])?)?(_Promo)?$`,
)

// unParsableVMSizes map holds vmSize strings that cannot be easily parsed with skuSizeScheme.
var unParsableVMSizes = map[string]VMSizeType{
	"M416s_8_v2": {
		family:                      "M",
		subfamily:                   nil,
		cpus:                        "416",
		cpusConstrained:             nil,
		additiveFeatures:            []rune{'s'},
		acceleratorType:             nil,
		confidentialChildCapability: false,
		version:                     "v2",
		promoVersion:                false,
		series:                      "Ms_v2",
	},
}

type VMSizeType struct {
	family                      string
	subfamily                   *string
	cpus                        string
	cpusConstrained             *string
	additiveFeatures            []rune
	acceleratorType             *string
	confidentialChildCapability bool
	version                     string
	promoVersion                bool
	series                      string
}

// parseVMSize parses the VM size and returns the parts as a map.
func parseVMSize(vmSizeName string) ([]string, error) {
	parts := skuSizeScheme.FindStringSubmatch(vmSizeName)
	if parts == nil || len(parts) < 10 {
		return nil, fmt.Errorf("could not parse VM size %s", vmSizeName)
	}
	return parts, nil
}

// getVMSize is a helper function used by GetVMSize() in sku.go
func getVMSize(vmSizeName string) (*VMSizeType, error) {
	vmSize := VMSizeType{}

	parts, err := parseVMSize(vmSizeName)
	if err != nil {
		if vmSizeVal, ok := unParsableVMSizes[vmSizeName]; ok {
			return &vmSizeVal, nil
		}
		return nil, err
	}

	// [Family] - ([A-Z]): Captures a single uppercase letter.
	vmSize.family = parts[1]

	// [Sub-family]* - ([A-Z]?): Optionally captures another uppercase letter.
	if len(parts[2]) > 0 {
		vmSize.subfamily = &parts[2]
	}

	// [# of vCPUs] - ([0-9]+): Captures one or more digits.
	vmSize.cpus = parts[3]

	// [Constrained vCPUs]*
	// -?: Optionally captures a hyphen.
	// ((?:[0-9]+)?): Optionally captures another sequence of one or more digits.
	if len(parts[4]) > 0 {
		_, err := strconv.Atoi(parts[4])
		if err != nil {
			return nil, fmt.Errorf("converting constrained CPUs, %w", err)
		}
		vmSize.cpusConstrained = &parts[4]
	}

	// [Additive Features]
	// ((?:[abcdilmtspPr]+|C+|NP)?): Captures a sequence of letters representing certain attributes.
	// It can capture combinations like 'abcdilmtspPr' or 'C+' or 'NP'.
	vmSize.additiveFeatures = []rune(parts[5])

	// [Accelerator Type]*
	// _?: Optionally captures an underscore.
	// (?:([A-Z][0-9]+)_?)?: Optionally captures a pattern that starts with an uppercase letter followed by digits,
	// followed by an optional underscore.
	if len(parts[6]) > 0 {
		vmSize.acceleratorType = &parts[6]
	}

	// [Confidential Child Capability]* - only AKS
	// (_cc_)?: Optionally captures the string "cc" with underscores on both sides.
	if parts[7] == "_cc_" {
		vmSize.confidentialChildCapability = true
	}

	// [Version]*
	// Optionally captures the pattern 'v' or 'V' followed by a digit from 1 to 9.
	vmSize.version = parts[8]

	// [Promo]*
	// (_Promo)?: Optionally captures the string "_Promo".
	if parts[9] == "_Promo" {
		vmSize.promoVersion = true
	}

	// [Series]
	subfamily := ""
	if vmSize.subfamily != nil {
		subfamily = *vmSize.subfamily
	}
	version := ""
	if len(vmSize.version) > 0 {
		version = "_" + vmSize.version
	}
	vmSize.series = vmSize.family + subfamily + string(vmSize.additiveFeatures) + version

	return &vmSize, nil
}
