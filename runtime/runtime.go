package devopsruntime

import (
	"encoding/json"
	log "github.com/EntropyPool/entropy-logger"
	machspec "github.com/EntropyPool/machine-spec"
	"github.com/jaypipes/ghw"
	"github.com/rai-project/nvidia-smi"
	"strings"
)

func rootInDisk(disk *ghw.Disk) bool {
	for _, part := range disk.Partitions {
		if part.MountPoint == "/" {
			return true
		}
	}
	return false
}

func getNvmeList() []string {
	block, _ := ghw.Block()

	nvmes := []string{}
	for _, disk := range block.Disks {
		if rootInDisk(disk) {
			continue
		}
		if strings.Contains(disk.Name, "nvme") {
			nvmes = append(nvmes, disk.Name)
		}
	}

	return nvmes
}

func GetNvmeCount() (int, error) {
	return len(getNvmeList()), nil
}

type diskInfo struct {
	Name   string `json:"name"`
	Vendor string `json:"vendor"`
	Model  string `json:"model"`
	Sn     string `json:"sn"`
	Wwn    string `json:"wwn"`
}

func GetNvmeDesc() ([]string, error) {
	nvmeDescs := []string{}

	block, _ := ghw.Block()
	for _, disk := range block.Disks {
		if rootInDisk(disk) {
			continue
		}
		if !strings.Contains(disk.Name, "nvme") {
			continue
		}
		info := diskInfo{
			Name:   disk.Name,
			Vendor: disk.Vendor,
			Model:  disk.Model,
			Sn:     disk.SerialNumber,
			Wwn:    disk.WWN,
		}
		b, _ := json.Marshal(&info)
		nvmeDescs = append(nvmeDescs, string(b))
	}

	return nvmeDescs, nil
}

type gpuDesc struct {
	Vendor        string `json:"vendor"`
	Product       string `json:"product"`
	DriverVersion string `json:"driver_version"`
}

func GetGpuCount() (int, error) {
	gpu, _ := ghw.GPU()
	return len(gpu.GraphicsCards), nil
}

func GetGpuDesc() ([]string, error) {
	gpus := []string{}

	nvgpu, err := nvidiasmi.New()
	if err == nil {
		for _, gpu := range nvgpu.GPUS {
			info := gpuDesc{
				Vendor:        "NVIDIA",
				Product:       gpu.ProductName,
				DriverVersion: nvgpu.DriverVersion,
			}
			desc, _ := json.Marshal(info)
			gpus = append(gpus, string(desc))
		}
	}

	gpu, _ := ghw.GPU()
	for _, card := range gpu.GraphicsCards {
		added := false
		for _, name := range gpus {
			if strings.Contains(name, card.DeviceInfo.Product.Name) {
				added = true
			}
		}
		if added {
			continue
		}
		info := gpuDesc{
			Vendor:  card.DeviceInfo.Vendor.Name,
			Product: card.DeviceInfo.Product.Name,
		}
		desc, _ := json.Marshal(info)
		gpus = append(gpus, string(desc))
	}

	return gpus, nil
}

func GetMemoryCount() (int, error) {
	memory, _ := ghw.Memory()

	if memory.Modules == nil || len(memory.Modules) == 0 {
		spec := machspec.NewMachineSpec()
		err := spec.PrepareLowLevel()
		if err != nil {
			log.Errorf(log.Fields{}, "fail to prepare spec: %v", err)
		}
		return len(spec.Memory), nil
	} else {
		return len(memory.Modules), nil
	}

	return 0, nil
}

func GetMemorySize() (uint64, error) {
	memory, _ := ghw.Memory()
	return uint64(memory.TotalPhysicalBytes), nil
}

func GetMemoryDesc() ([]string, error) {
	mems := []string{}

	memory, _ := ghw.Memory()
	for _, m := range memory.Modules {
		info := machspec.Memory{
			Dimm:         m.Location,
			Sn:           m.SerialNumber,
			SizeGB:       int(m.SizeBytes / 1024 / 1024 / 1024),
			Manufacturer: m.Vendor,
		}
		b, _ := json.Marshal(info)
		mems = append(mems, string(b))
	}

	if len(mems) == 0 {
		spec := machspec.NewMachineSpec()
		err := spec.PrepareLowLevel()
		if err == nil {
			for _, m := range spec.Memory {
				b, _ := json.Marshal(m)
				mems = append(mems, string(b))
			}
		}
	}

	return mems, nil
}
