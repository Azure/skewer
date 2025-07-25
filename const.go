package skewer

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
	// GPUs identifies the capability for the number of GPUS.
	GPUs = "GPUs"
	// MemoryGB identifies the capability for memory capacity.
	MemoryGB = "MemoryGB"
	// HyperVGenerations identifies the hyper-v generations this vm sku supports.
	HyperVGenerations = "HyperVGenerations"
	// EncryptionAtHost identifies the capability for accelerated networking support.
	EncryptionAtHost = "EncryptionAtHostSupported"
	// UltraSSDAvailable identifies the capability for ultra ssd
	// enablement.
	UltraSSDAvailable = "UltraSSDAvailable"
	// CachedDiskBytes identifies the maximum size of the cache disk for
	// a vm.
	CachedDiskBytes = "CachedDiskBytes"
	// MaxResourceVolumeMB identifies the maximum size of the temporary
	// disk for a vm.
	MaxResourceVolumeMB = "MaxResourceVolumeMB"
	// CapabilityPremiumIO identifies the capability for PremiumIO.
	CapabilityPremiumIO = "PremiumIO"
	// CapabilityCpuArchitectureType identifies the type of CPU architecture (x64,Arm64).
	CapabilityCPUArchitectureType = "CpuArchitectureType"
	// CapabilityTrustedLaunchDisabled identifes whether TrustedLaunch is disabled.
	CapabilityTrustedLaunchDisabled = "TrustedLaunchDisabled"
	// CapabilityConfidentialComputingType identifies the type of ConfidentialComputing.
	CapabilityConfidentialComputingType = "ConfidentialComputingType"
	// ConfidentialComputingTypeSNP denoted the "SNP" ConfidentialComputing.
	ConfidentialComputingTypeSNP = "SNP"
	// DiskControllerTypes identifies the disk controller types supported by the VM SKU.
	DiskControllerTypes = "DiskControllerTypes"
	// SupportedEphemeralOSDiskPlacements identifies supported ephemeral OS disk placements.
	SupportedEphemeralOSDiskPlacements = "SupportedEphemeralOSDiskPlacements"
	// NvmeDiskSizeInMiB identifies the NVMe disk size in MiB.
	NvmeDiskSizeInMiB = "NvmeDiskSizeInMiB"
)

const (
	// HyperVGeneration1 identifies a sku which supports HyperV
	// Generation 1.
	HyperVGeneration1 = "V1"
	// HyperVGeneration2 identifies a sku which supports HyperV
	// Generation 2.
	HyperVGeneration2 = "V2"
)

const (
	// DiskControllerSCSI identifies the SCSI disk controller type.
	DiskControllerSCSI = "SCSI"
	// DiskControllerNVMe identifies the NVMe disk controller type.
	DiskControllerNVMe = "NVMe"
	// EphemeralDiskPlacementNvme identifies NVMe disk placement for ephemeral OS disk.
	EphemeralDiskPlacementNvme = "NvmeDisk"
)

const (
	ten       = 10
	sixtyFour = 64
)
