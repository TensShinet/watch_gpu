package watch

import (
	"log"

	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
)

var DeviceCount int
var Devices []*nvml.Device
var HostName string

// nvml.ProcessInfo
// type ProcessInfo struct {
// 	PID        uint
// 	Name       string
// 	MemoryUsed uint64
// 	Type       ProcessType
// }

type GpuProcess struct {
	GPU         int
	PID         int
	Name        string
	MemoryUsed  uint64
	Type        uint
	MemoryUsage int
}

func GetAllRunningProcesses() (gpuProcesses []GpuProcess, err error) {
	nvml.Init()
	defer nvml.Shutdown()
	DeviceCount, err := nvml.GetDeviceCount()
	if err != nil {
		log.Panicln("Error getting device count:", err)
	}

	var Devices []*nvml.Device
	for i := uint(0); i < DeviceCount; i++ {
		device, err := nvml.NewDevice(i)
		if err != nil {
			log.Panicf("Error getting device %d: %v\n", i, err)
		}
		Devices = append(Devices, device)
	}
	for i, device := range Devices {
		pInfo, err := device.GetAllRunningProcesses()
		if err != nil {
			log.Panicf("Error getting device %d processes: %v\n", i, err)
		}
		for j := range pInfo {
			gpuProcess := GpuProcess{
				i,
				int(pInfo[j].PID),
				pInfo[j].Name,
				pInfo[j].MemoryUsed,
				uint(pInfo[j].Type),
				int(float32(pInfo[j].MemoryUsed) / float32(*device.Memory) * 100),
			}
			gpuProcesses = append(gpuProcesses, gpuProcess)
		}
	}

	return gpuProcesses, nil
}
