package skewer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/pkg/errors"
)

// SKU wraps an Azure compute SKU with richer functionality
type SKU compute.ResourceSku

// Wrap takes an array of compute resource skus and wraps them into an
// array of our richer type.
func Wrap(in []compute.ResourceSku) []SKU {
	out := make([]SKU, len(in))
	for index, value := range in {
		out[index] = SKU(value)
	}
	return out
}

const (
	// VirtualMachines is the .
	VirtualMachines = "virtualMachines"
	// Disks is a convenience constant to filter resource SKUs to only include disks.
	Disks = "disks"
)

// Supported models an enum of possible boolean values for resource support in the Azure API.
type Supported string

const (
	// CapabilitySupported is an enum value for the string "True" returned when a SKU supports a binary capability.
	CapabilitySupported Supported = "True"
	// CapabilityUnsupported is an enum value for the string "False" returned when a SKU does not support a binary capability.
	CapabilityUnsupported Supported = "False"
)

const (
	// EphemeralOSDisk identifies the capability for ephemeral os support.
	EphemeralOSDisk = "EphemeralOSDiskSupported"
	// AcceleratedNetworking identifies the capability for accelerated networking support.
	AcceleratedNetworking = "AcceleratedNetworkingEnabled"
	// VCPUs identifies the capability for the number of vCPUS.
	VCPUs = "vCPUs"
	// MemoryGB identifies the capability for memory capacity.
	MemoryGB = "MemoryGB"
	// HyperVGenerations identifies the hyper-v generations this vm sku supports.
	HyperVGenerations = "HyperVGenerations"
	// EncryptionAtHost identifies the capability for accelerated networking support.
	EncryptionAtHost = "EncryptionAtHostSupported"
	// UltraSSDAvailable identifies the capability for ultra ssd
	// enablement.
	UltraSSDAvailable = "UltraSSDAvailable"
	// CachedDiskBytes identifies the maximum size of the cach disk for
	// a vm.
	CachedDiskBytes = "CachedDiskBytes"
)

// ErrCapabilityNotFound will be returned when a capability could not be
// found, even without a value.
type ErrCapabilityNotFound struct {
	capability string
}

func (e *ErrCapabilityNotFound) Error() string {
	return e.capability + "CapabilityNotFound"
}

// ErrCapabilityValueNil will be returned when a capability was found by
// name but the value was nil.
type ErrCapabilityValueNil struct {
	capability string
}

func (e *ErrCapabilityValueNil) Error() string {
	return e.capability + "CapabilityValueNil"
}

// ErrCapabilityValueParse will be returned when a capability was found by
// name but the value was nil.
type ErrCapabilityValueParse struct {
	capability string
	value      string
	err        error
}

func (e *ErrCapabilityValueParse) Error() string {
	return fmt.Sprintf("%sCapabilityValueParse: failed to parse string '%s' as int64, error: '%s'", e.capability, e.value, e.err)
}

// VCPU returns the number of vCPUs this SKU supports.
func (s *SKU) VCPU() (int64, error) {
	return s.GetCapabilityIntegerQuantity(VCPUs)
}

// Memory returns the amount of memory this SKU supports.
func (s *SKU) Memory() (float64, error) {
	return s.GetCapabilityFloatQuantity(MemoryGB)
}

func (s *SKU) MaxCachedDiskBytes() (int64, error) {
	return s.GetCapabilityIntegerQuantity(CachedDiskBytes)
}

func (s *SKU) IsEncryptionAtHostSupported() bool {
	return s.HasCapability(EncryptionAtHost)
}

func (s *SKU) IsUltraSSDAvailable() bool {
	return s.HasZonalCapability(UltraSSDAvailable)
}

func (s *SKU) IsEphemeralOSDiskSupported() bool {
	return s.HasCapability(EphemeralOSDisk)
}

// GetCapabilityIntegerQuantity retrieves and parses the value of an
// integer numeric capability with the provided name. It errors if the
// capability is not found, the value was nil, or the value could not be
// parsed as an integer.
func (s *SKU) GetCapabilityIntegerQuantity(name string) (int64, error) {
	if s.Capabilities == nil {
		return -1, &ErrCapabilityNotFound{name}
	}
	for _, capability := range *s.Capabilities {
		if capability.Name != nil && *capability.Name == name {
			if capability.Value != nil {
				intVal, err := strconv.ParseInt(*capability.Value, 10, 64)
				if err != nil {
					return -1, &ErrCapabilityValueParse{name, *capability.Value, err}
				}
				return intVal, nil
			}
			return -1, &ErrCapabilityValueNil{name}
		}
	}
	return -1, &ErrCapabilityNotFound{name}
}

// GetCapabilityFloatQuantity retrieves and parses the value of a
// floating point numeric capability with the provided name. It errors
// if the capability is not found, the value was nil, or the value could
// not be parsed as an integer.
func (s *SKU) GetCapabilityFloatQuantity(name string) (float64, error) {
	if s.Capabilities == nil {
		return -1, &ErrCapabilityNotFound{name}
	}
	for _, capability := range *s.Capabilities {
		if capability.Name != nil && *capability.Name == name {
			if capability.Value != nil {
				intVal, err := strconv.ParseFloat(*capability.Value, 64)
				if err != nil {
					return -1, &ErrCapabilityValueParse{name, *capability.Value, err}
				}
				return intVal, nil
			}
			return -1, &ErrCapabilityValueNil{name}
		}
	}
	return -1, &ErrCapabilityNotFound{name}
}

// HasCapability return true for a capability which can be either
// supported or not. Examples include "EphemeralOSDiskSupported",
// "EncryptionAtHostSupported", "AcceleratedNetworkingEnabled", and
// "RdmaEnabled"
func (s *SKU) HasCapability(name string) bool {
	if s.Capabilities == nil {
		return false
	}
	for _, capability := range *s.Capabilities {
		if capability.Name != nil && strings.EqualFold(*capability.Name, name) {
			return capability.Value != nil && strings.EqualFold(*capability.Value, string(CapabilitySupported))
		}
	}
	return false
}

// HasZonalCapability return true for a capability which can be either
// supported or not. Examples include "UltraSSDAvailable".
// This function only checks that zone details suggest support: it will
// return true for a whole location even when only one zone supports the
// feature. Currently, the only real scenario that appears to use
// zoneDetails is UltraSSDAvailable which always lists all regions as
// available.
// TODO(ace): update this function signature/behavior if necessary to
// account for per-zone availability.
func (s *SKU) HasZonalCapability(name string) bool {
	if s.LocationInfo == nil {
		return false
	}
	for _, locationInfo := range *s.LocationInfo {
		if locationInfo.ZoneDetails == nil {
			continue
		}
		for _, zoneDetails := range *locationInfo.ZoneDetails {
			if zoneDetails.Capabilities == nil {
				continue
			}
			for _, capability := range *zoneDetails.Capabilities {
				if capability.Name != nil && strings.EqualFold(*capability.Name, name) {
					if capability.Value != nil && strings.EqualFold(*capability.Value, string(CapabilitySupported)) {
						return true
					}
				}
			}
		}
	}
	return false
}

// HasCapabilityWithSeparator return true for a capability which may be
// exposed as a comma-separated list. We check that the list contains
// the desired substring. An example is "HyperVGenerations" which may be
// "V1,V2"
func (s *SKU) HasCapabilityWithSeparator(name, value string) bool {
	if s.Capabilities == nil {
		return false
	}
	for _, capability := range *s.Capabilities {
		if capability.Name != nil && strings.EqualFold(*capability.Name, name) {
			return capability.Value != nil && strings.Contains(*capability.Value, value)
		}
	}
	return false
}

// HasCapabilityWithCapacity returns true when the provided resource
// exposes a numeric capability and the maximum value exposed by that
// capability exceeds the value requested by the user. Examples include
// "MaxResourceVolumeMB", "OSVhdSizeMB", "vCPUs",
// "MemoryGB","MaxDataDiskCount", "CombinedTempDiskAndCachedIOPS",
// "CombinedTempDiskAndCachedReadBytesPerSecond",
// "CombinedTempDiskAndCachedWriteBytesPerSecond", "UncachedDiskIOPS",
// and "UncachedDiskBytesPerSecond"
func (s *SKU) HasCapabilityWithCapacity(name string, value int64) (bool, error) {
	if s.Capabilities == nil {
		return false, nil
	}
	for _, capability := range *s.Capabilities {
		if capability.Name != nil && strings.EqualFold(*capability.Name, name) {
			if capability.Value != nil {
				intVal, err := strconv.ParseInt(*capability.Value, 10, 64)
				if err != nil {
					return false, errors.Wrapf(err, "failed to parse string '%s' as int64", *capability.Value)
				}
				if intVal >= value {
					return true, nil
				}
			}
			return false, nil
		}
	}
	return false, nil
}

// IsAvailable returns true when the requested location matches one on
// the sku, and there are no total restrictions on the location.
func (s *SKU) IsAvailable(location string) bool {
	if s.LocationInfo == nil {
		return false
	}
	for _, locationInfo := range *s.LocationInfo {
		if strings.EqualFold(*locationInfo.Location, location) {
			if s.Restrictions != nil {
				for _, restriction := range *s.Restrictions {
					// Can't deploy to any zones in this location. We're done.
					if restriction.Type == compute.Location {
						return false
					}
				}
			}
			return true
		}
	}
	return false
}

// IsRestricted returns true when a location restriction exists for
// this SKU.
func (s *SKU) IsRestricted(location string) bool {
	if s.Restrictions == nil {
		return false
	}
	for _, restriction := range *s.Restrictions {
		if restriction.Values == nil {
			continue
		}
		for _, candidate := range *restriction.Values {
			// Can't deploy in this location. We're done.
			if strings.EqualFold(candidate, location) && restriction.Type == compute.Location {
				return true
			}
		}
	}
	return false
}

// IsResourceType returns true when the wrapped SKU has the provided
// value as its resource type. This may be used to filter using values
// such as "virtualMachines", "disks", "availabilitySets", "snapshots",
// and "hostGroups/hosts".
func (s *SKU) IsResourceType(t string) bool {
	return s.ResourceType != nil && strings.EqualFold(*s.ResourceType, t)
}

// GetResourceType returns the name of this resource sku. It normalizes pointers
// to the empty string for comparison purposes. For example,
// "virtualMachines" for a virtual machine.
func (s *SKU) GetResourceType() string {
	if s.ResourceType == nil {
		return ""
	}
	return *s.ResourceType
}

// GetName returns the name of this resource sku. It normalizes pointers
// to the empty string for comparison purposes. For example,
// "Standard_D8s_v3" for a virtual machine.
func (s *SKU) GetName() string {
	if s.Name == nil {
		return ""
	}
	return *s.Name
}

// GetLocation returns the first found location on this *SKU resource.
// Typically only one should be listed (multiple SKU results will be returned for multiple regions).
// We fallback to locationInfo although this appears to be duplicate info.
func (s *SKU) GetLocation() string {
	if s.Locations != nil {
		for _, location := range *s.Locations {
			return location
		}
	}

	// TODO(ace): probably should remove
	if s.LocationInfo != nil {
		for _, locationInfo := range *s.LocationInfo {
			if locationInfo.Location != nil {
				return *locationInfo.Location
			}
		}
	}

	return ""
}

// AvailabilityZones returns the list of Availability Zones which have this resource SKU available and unrestricted.
func (s *SKU) AvailabilityZones(location string) map[string]bool {
	for _, locationInfo := range *s.LocationInfo {
		if strings.EqualFold(*locationInfo.Location, location) {
			// Use map for easy deletion and iteration
			availableZones := make(map[string]bool)

			// add all zones
			for _, zone := range *locationInfo.Zones {
				availableZones[zone] = true
			}

			if s.Restrictions != nil {
				for _, restriction := range *s.Restrictions {
					// Can't deploy to any zones in this location. We're done.
					if restriction.Type == compute.Location {
						availableZones = nil
						break
					}

					// remove restricted zones
					for _, restrictedZone := range *restriction.RestrictionInfo.Zones {
						delete(availableZones, restrictedZone)
					}
				}
			}

			return availableZones
		}
	}

	return nil
}

// Equal returns true when two skus have the same location, type, and name.
func (s *SKU) Equal(other *SKU) bool {
	return strings.EqualFold(s.GetResourceType(), other.GetResourceType()) &&
		strings.EqualFold(s.GetName(), other.GetName()) &&
		strings.EqualFold(s.GetLocation(), other.GetLocation())
}
