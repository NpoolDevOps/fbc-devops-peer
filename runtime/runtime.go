package devopsruntime

import (
	_ "github.com/EntropyPool/entropy-logger"
	"github.com/shirou/gopsutil/disk"
	"strings"
)

type diskInfo struct {
	name        string `json:"disk_name"`
	description string `json:"disk_discription"`
}

func getRootPart() string {
	infos, _ := disk.Partitions(false)
	for _, info := range infos {
		if info.Mountpoint == "/" {
			return info.Device
		}
	}
	return ""
}

func getNvmeList() []string {
	infos, _ := disk.Partitions(false)

	nvmes := []string{}
	for _, info := range infos {
		if strings.Contains(info.Device, "nvme") {
			nvmes = append(nvmes, info.Device)
		}
	}

	return nvmes
}

func GetNvmeCount() (int, error) {
	nvmeList := getNvmeList()
	rootPart := getRootPart()

	nvmeCount := 0
	for _, nvme := range nvmeList {
		if strings.Contains(rootPart, nvme) {
			continue
		}
		nvmeCount += 1
	}

	return nvmeCount, nil
}

func GetNvmeDesc() ([]string, error) {
	return nil, nil
}

func GetGpuCount() (int, error) {
	return 0, nil
}

func GetGpuDesc() ([]string, error) {
	return nil, nil
}
