package dto

type GPUList struct {
	DriverVersion     string
	CUDADriverVersion string
	Count             int
	GPUInfo           []*GPUStatus
}

type GPUStatus struct {
	DeviceName        string
	UUID              string
	MemoryUsed        int
	MemoryTotal       int
	PowerUsage        int
	PowerDefaultLimit int
	ProcessInfo       []*ProcessInfoCombined
}

type ProcessInfoCombined struct {
	Pid           int
	CommandLine   string
	User          string
	UsedGpuMemory int
}

type CPUList struct {
	Count   int
	CPUInfo []*CPUStatus
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
