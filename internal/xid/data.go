package xid

// XidInfo contains information about an NVIDIA Xid error
type XidInfo struct {
	Name        string   // Short name of the error
	Severity    string   // "Critical", "Warning", "Info"
	Description string   // Detailed description
	Causes      []string // Common causes
	Actions     []string // Recommended actions
	Reference   string   // Reference URL
}

// Database contains information about known Xid errors
var Database = map[int]XidInfo{
	8: {
		Name:        "GPU Initialization Error",
		Severity:    "Critical",
		Description: "GPU failed to initialize properly. This may indicate driver or hardware issues.",
		Causes: []string{
			"Driver installation issue",
			"GPU hardware failure",
			"BIOS/UEFI misconfiguration",
		},
		Actions: []string{
			"Reinstall NVIDIA driver",
			"Check GPU seating in PCIe slot",
			"Update system BIOS",
		},
		Reference: "https://docs.nvidia.com/deploy/xid-errors/",
	},
	13: {
		Name:        "Graphics Engine Exception",
		Severity:    "Critical",
		Description: "GPU encountered an unrecoverable graphics engine error.",
		Causes: []string{
			"Driver bug",
			"GPU hardware failure",
			"Overclocking instability",
		},
		Actions: []string{
			"Check GPU temperature",
			"Update NVIDIA driver",
			"Run GPU diagnostics (nvidia-smi -q)",
			"If overclocked, reduce clock speeds",
		},
		Reference: "https://docs.nvidia.com/deploy/xid-errors/",
	},
	31: {
		Name:        "GPU Memory Page Fault",
		Severity:    "Critical",
		Description: "GPU detected invalid memory access. Process may have accessed invalid GPU memory.",
		Causes: []string{
			"Application bug (illegal memory access)",
			"GPU memory corruption",
			"Driver issue",
		},
		Actions: []string{
			"Check application code for memory errors",
			"Run GPU memory test (nvidia-smi --gpu-reset)",
			"Update driver",
			"Check for ECC errors",
		},
		Reference: "https://docs.nvidia.com/deploy/xid-errors/",
	},
	32: {
		Name:        "Invalid or Corrupted Push Buffer",
		Severity:    "Critical",
		Description: "GPU received corrupted command data.",
		Causes: []string{
			"Driver bug",
			"System memory corruption",
			"PCIe bus errors",
		},
		Actions: []string{
			"Update NVIDIA driver",
			"Check system memory with memtest86+",
			"Check PCIe slot and cable connections",
		},
		Reference: "https://docs.nvidia.com/deploy/xid-errors/",
	},
	43: {
		Name:        "GPU Stopped Processing",
		Severity:    "Critical",
		Description: "GPU stopped responding to commands.",
		Causes: []string{
			"Infinite loop in shader/kernel",
			"Hardware hang",
			"Driver timeout",
		},
		Actions: []string{
			"Check for long-running GPU kernels",
			"Increase TDR timeout if needed",
			"Update NVIDIA driver",
		},
		Reference: "https://docs.nvidia.com/deploy/xid-errors/",
	},
	45: {
		Name:        "Preemptive Cleanup",
		Severity:    "Warning",
		Description: "GPU driver performed preemptive cleanup due to long-running operation.",
		Causes: []string{
			"Long-running GPU kernel",
			"Resource contention",
		},
		Actions: []string{
			"Optimize GPU kernel execution time",
			"Check for deadlocks in application",
		},
		Reference: "https://docs.nvidia.com/deploy/xid-errors/",
	},
	48: {
		Name:        "Double Bit ECC Error",
		Severity:    "Critical",
		Description: "Uncorrectable ECC memory error detected. Data corruption has occurred.",
		Causes: []string{
			"GPU memory hardware failure",
			"Memory aging",
			"Cosmic ray bit flip (rare)",
		},
		Actions: []string{
			"Immediately drain workloads from node",
			"Check ECC error counts with nvidia-smi",
			"Schedule GPU replacement",
			"Reset GPU with nvidia-smi --gpu-reset",
		},
		Reference: "https://docs.nvidia.com/deploy/xid-errors/",
	},
	56: {
		Name:        "Display Engine Error",
		Severity:    "Warning",
		Description: "Error in display engine. Usually not critical for compute workloads.",
		Causes: []string{
			"Display driver issue",
			"Monitor connection problem",
		},
		Actions: []string{
			"If headless server, can usually be ignored",
			"Update display driver",
		},
		Reference: "https://docs.nvidia.com/deploy/xid-errors/",
	},
	61: {
		Name:        "Internal Micro-controller Halt",
		Severity:    "Critical",
		Description: "GPU internal processor halted unexpectedly.",
		Causes: []string{
			"Firmware bug",
			"Hardware failure",
		},
		Actions: []string{
			"Restart the system",
			"Update GPU firmware",
			"Contact NVIDIA support if recurring",
		},
		Reference: "https://docs.nvidia.com/deploy/xid-errors/",
	},
	62: {
		Name:        "ECC Memory Page Retired",
		Severity:    "Warning",
		Description: "GPU retired a memory page due to ECC errors. Memory capacity slightly reduced.",
		Causes: []string{
			"Memory cell wear",
			"Manufacturing defect",
		},
		Actions: []string{
			"Monitor retired page count with nvidia-smi",
			"Plan GPU replacement if count increases rapidly",
		},
		Reference: "https://docs.nvidia.com/deploy/xid-errors/",
	},
	63: {
		Name:        "ECC Page Retirement: Row Remapping Failure",
		Severity:    "Warning",
		Description: "GPU retired memory pages due to ECC errors, but row remapping failed.",
		Causes: []string{
			"Memory cell wear",
			"Manufacturing defect",
			"Too many retired pages",
		},
		Actions: []string{
			"Monitor ECC error trends",
			"Plan for GPU replacement if errors increase",
			"Consider reduced workload on affected GPU",
		},
		Reference: "https://docs.nvidia.com/deploy/xid-errors/",
	},
	64: {
		Name:        "ECC Page Retirement: High Error Rate",
		Severity:    "Warning",
		Description: "High rate of ECC errors causing page retirement.",
		Causes: []string{
			"Memory degradation",
			"Environmental factors (temperature, voltage)",
		},
		Actions: []string{
			"Check GPU temperature",
			"Monitor power delivery",
			"Plan GPU replacement",
		},
		Reference: "https://docs.nvidia.com/deploy/xid-errors/",
	},
	68: {
		Name:        "Video Processor Exception",
		Severity:    "Warning",
		Description: "Error in video processing engine.",
		Causes: []string{
			"Video codec issue",
			"Driver bug",
		},
		Actions: []string{
			"Update driver",
			"Check video encoding/decoding parameters",
		},
		Reference: "https://docs.nvidia.com/deploy/xid-errors/",
	},
	69: {
		Name:        "Graphics Engine Class Error",
		Severity:    "Critical",
		Description: "Class error in graphics engine.",
		Causes: []string{
			"Driver bug",
			"Hardware issue",
		},
		Actions: []string{
			"Update NVIDIA driver",
			"Run diagnostics",
		},
		Reference: "https://docs.nvidia.com/deploy/xid-errors/",
	},
	74: {
		Name:        "GPU Recovered with Reset",
		Severity:    "Warning",
		Description: "GPU was reset and recovered. Previous error caused temporary unavailability.",
		Causes: []string{
			"Transient hardware issue",
			"Driver recovery mechanism triggered",
		},
		Actions: []string{
			"Monitor for recurring issues",
			"Check previous Xid errors for root cause",
		},
		Reference: "https://docs.nvidia.com/deploy/xid-errors/",
	},
	79: {
		Name:        "GPU Fallen Off Bus",
		Severity:    "Critical",
		Description: "GPU disconnected from PCIe bus. System cannot communicate with GPU.",
		Causes: []string{
			"PCIe slot contact issue",
			"Power supply instability",
			"GPU hardware failure",
			"Thermal shutdown",
			"Driver/firmware bug",
		},
		Actions: []string{
			"Immediately drain node from workloads",
			"Check BMC/IPMI hardware event logs",
			"Check GPU temperature history",
			"Reseat GPU in PCIe slot",
			"Check power connections",
			"If recurring, replace GPU",
		},
		Reference: "https://docs.nvidia.com/deploy/xid-errors/",
	},
	92: {
		Name:        "High Single-bit ECC Error Rate",
		Severity:    "Warning",
		Description: "High rate of correctable ECC errors detected.",
		Causes: []string{
			"Memory degradation beginning",
			"Environmental factors",
		},
		Actions: []string{
			"Monitor error rate trend",
			"Check cooling and temperature",
			"Plan preventive maintenance",
		},
		Reference: "https://docs.nvidia.com/deploy/xid-errors/",
	},
	94: {
		Name:        "Contained ECC Error",
		Severity:    "Warning",
		Description: "ECC error was contained and corrected. No data corruption.",
		Causes: []string{
			"Memory cell degradation",
			"Cosmic ray bit flip",
		},
		Actions: []string{
			"Monitor for increasing frequency",
			"No immediate action needed if isolated",
			"Track trends over time",
		},
		Reference: "https://docs.nvidia.com/deploy/xid-errors/",
	},
	95: {
		Name:        "Uncontained ECC Error",
		Severity:    "Critical",
		Description: "ECC error could not be contained. Potential data corruption.",
		Causes: []string{
			"Severe memory failure",
			"Multiple simultaneous bit errors",
		},
		Actions: []string{
			"Stop all workloads on affected GPU",
			"Check job outputs for corruption",
			"Schedule GPU replacement",
		},
		Reference: "https://docs.nvidia.com/deploy/xid-errors/",
	},
	119: {
		Name:        "GSP Failure",
		Severity:    "Critical",
		Description: "GPU System Processor (GSP) encountered a failure.",
		Causes: []string{
			"Firmware issue",
			"Hardware failure",
		},
		Actions: []string{
			"Reset GPU with nvidia-smi --gpu-reset",
			"Update GPU firmware",
			"Restart system if reset fails",
		},
		Reference: "https://docs.nvidia.com/deploy/xid-errors/",
	},
}

// GetXidInfo returns information about a specific Xid error
func GetXidInfo(code int) (XidInfo, bool) {
	info, ok := Database[code]
	return info, ok
}

// ListAllXids returns all known Xid codes
func ListAllXids() []int {
	codes := make([]int, 0, len(Database))
	for code := range Database {
		codes = append(codes, code)
	}
	return codes
}
