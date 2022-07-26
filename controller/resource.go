package controller

import (
	"encoding/json"
	"errors"
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/gin-gonic/gin"
	"github.com/joexu01/container-dispatcher/dto"
	"github.com/joexu01/container-dispatcher/middleware"
	"github.com/shirou/gopsutil/v3/process"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

type ResourceController struct{}

func ResourceControllerRegister(group *gin.RouterGroup) {
	res := &ResourceController{}
	group.GET("/gpu", res.ResourceGPUList)
	group.GET("/cpu", res.ResourceCPUList)
	//group.GET("/gpu/proc")
}

// ResourceGPUList godoc
// @Summary      获取GPU基本的状态
// @Description  获取GPU基本的状态包含个数、占用率等
// @Tags         resource
// @Produce      json
// @Success      200  {object}  middleware.Response{data=dto.GPUList} "success"
// @Failure      500  {object}  middleware.Response
// @Router       /resource/gpu [get]
func (r *ResourceController) ResourceGPUList(c *gin.Context) {
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2000, errors.New("failed to initialize nvidia-smi sdk"), "")
		return
	}
	defer func() {
		_ = nvml.Shutdown()
	}()

	driverVer, _ := nvml.SystemGetDriverVersion()
	cudaDriverVerRaw, _ := nvml.SystemGetCudaDriverVersion()

	iToA := strconv.Itoa(cudaDriverVerRaw)
	split := strings.Split(iToA, "0")
	var cudaDriverVer string
	if len(split) >= 2 {
		cudaDriverVer = split[0] + "." + split[1]
	} else {
		cudaDriverVer = ""
	}

	var GpuInfoSlice []*dto.GPUStatus
	count, _ := nvml.DeviceGetCount()

	for gpuIdx := 0; gpuIdx < count; gpuIdx++ {
		deviceHandler, _ := nvml.DeviceGetHandleByIndex(gpuIdx)
		uuid, _ := deviceHandler.GetUUID()
		name, _ := deviceHandler.GetName()
		memory, _ := deviceHandler.GetMemoryInfo()
		powerUsage, _ := deviceHandler.GetPowerUsage()
		powerLimit, _ := deviceHandler.GetPowerManagementDefaultLimit()

		runningProc, _ := deviceHandler.GetGraphicsRunningProcesses()
		var procInfoSlice []*dto.ProcessInfoCombined
		for _, proc := range runningProc {
			procInst := proc

			p, _ := process.NewProcess(int32(int(procInst.Pid)))
			username, _ := p.Username()
			cmdLine, _ := p.Cmdline()

			procInfo := &dto.ProcessInfoCombined{
				Pid:           int(procInst.Pid),
				CommandLine:   cmdLine,
				User:          username,
				UsedGpuMemory: int(procInst.UsedGpuMemory / (1024 * 1024)),
			}
			procInfoSlice = append(procInfoSlice, procInfo)
		}

		status := &dto.GPUStatus{
			DeviceName:        name,
			UUID:              uuid,
			MemoryUsed:        int(memory.Used / (1024 * 1024)),
			MemoryTotal:       int(memory.Total / (1024 * 1024)),
			PowerUsage:        int(powerUsage / 1000),
			PowerDefaultLimit: int(powerLimit / 1000),
			ProcessInfo:       procInfoSlice,
		}
		GpuInfoSlice = append(GpuInfoSlice, status)
	}

	gpu := &dto.GPUList{
		DriverVersion:     driverVer,
		CUDADriverVersion: cudaDriverVer,
		Count:             count,
		GPUInfo:           GpuInfoSlice,
	}

	middleware.ResponseSuccess(c, gpu)
}

// ResourceCPUList godoc
// @Summary      获取CPU基本的状态
// @Description  获取CPU基本的状态
// @Tags         resource
// @Produce      json
// @Success      200  {object}  middleware.Response "success"
// @Failure      500  {object}  middleware.Response
// @Router       /resource/cpu [get]
func (r *ResourceController) ResourceCPUList(c *gin.Context) {
	cmd := exec.Command("lscpu", "-J")
	bytes, err := cmd.Output()
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2000, err, "")
		return
	}
	//str := string(bytes)
	//str = strings.Replace(str, "\n", "", -1)
	//str = strings.Replace(str, "\t", "", -1)
	//str = strings.Replace(str, "\\", "", -1)
	rawInfo := &dto.LSCPUResult{}
	err = json.Unmarshal(bytes, rawInfo)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2001, err, "")
		return
	}

	middleware.ResponseSuccess(c, rawInfo)
}
