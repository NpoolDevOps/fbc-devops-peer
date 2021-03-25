package devopsruntime

import (
	"encoding/json"
	log "github.com/EntropyPool/entropy-logger"
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
	Vendor  string `json:"vendor"`
	Product string `json:"product"`
}

func GetGpuCount() (int, error) {
	gpu, _ := ghw.GPU()
	return len(gpu.GraphicsCards), nil
}

func GetGpuDesc() ([]string, error) {
	gpu, _ := ghw.GPU()

	gpus := []string{}
	for _, card := range gpu.GraphicsCards {
		info := gpuDesc{
			Vendor:  card.DeviceInfo.Vendor.Name,
			Product: card.DeviceInfo.Product.Name,
		}
		desc, _ := json.Marshal(info)
		gpus = append(gpus, string(desc))
	}

	nvgpu, err := nvidiasmi.New()
	if err != nil {
		log.Infof(log.Fields{}, "==> %v", err)
	}
	log.Infof(log.Fields{}, "--> %v", nvgpu)

	return gpus, nil
}
