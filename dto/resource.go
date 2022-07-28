package dto

type GPUList struct {
	DriverVersion     string       `json:"driver_version"`
	CUDADriverVersion string       `json:"cuda_driver_version"`
	Count             int          `json:"count"`
	GPUInfo           []*GPUStatus `json:"gpu_info"`
}

type GPUStatus struct {
	DeviceName        string                 `json:"device_name"`
	UUID              string                 `json:"uuid"`
	MemoryUsed        int                    `json:"memory_used"`
	MemoryTotal       int                    `json:"memory_total"`
	PowerUsage        int                    `json:"power_usage"`
	PowerDefaultLimit int                    `json:"power_default_limit"`
	ProcessInfo       []*ProcessInfoCombined `json:"process_info"`
}

type ProcessInfoCombined struct {
	Pid           int    `json:"pid"`
	CommandLine   string `json:"command_line"`
	User          string `json:"user"`
	UsedGpuMemory int    `json:"used_gpu_memory"`
}

type CPUList struct {
	Count   int          `json:"count"`
	CPUInfo []*CPUStatus `json:"cpu_info"`
}

type CPUStatus struct {
	ModelName string
	VendorID  string
	MaxTurbo  float64
	Cores     int32
}

type LSCPUResult struct {
	Data []*LSCPUData `json:"lscpu"`
}

type LSCPUData struct {
	Field string `json:"field"`
	Data  string `json:"data"`
}
