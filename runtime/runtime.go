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

func GetGpuCount() (int, error) {
	return 0, nil
}

func GetGpuDesc() ([]string, error) {
	return nil, nil
}
