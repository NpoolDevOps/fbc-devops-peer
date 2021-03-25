package devopsruntime

import (
	"encoding/json"
	_ "github.com/EntropyPool/entropy-logger"
	"github.com/jaypipes/ghw"
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
	name   string `json:"name"`
	vendor string `json:"vendor"`
	model  string `json:"model"`
	sn     string `json:"sn"`
	wwn    string `json:"wwn"`
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
			name:   disk.Name,
			vendor: disk.Vendor,
			model:  disk.Model,
			sn:     disk.SerialNumber,
			wwn:    disk.WWN,
		}
		b, _ := json.Marshal(info)
		nvmeDescs = append(nvmeDescs, string(b))
	}

	return nvmeDescs, nil
}

func GetGpuCount() (int, error) {
	return 0, nil
}

func GetGpuDesc() ([]string, error) {
	return nil, nil
}
