package devopsruntime

import (
	"os/exec"
	"strings"
)

type diskInfo struct {
	name        string `json:"disk_name"`
	description string `json:"disk_discription"`
}

func getRootPart() string {
	out, _ := exec.Command("mount | grep \"on / type\" | awk '{print $1}'").Output()
	return string(out)
}

func getNvmeDiskList() []string {
	out, _ := exec.Command("lsblk | grep \"disk\" | grep \"nvme\"").Output()
	return strings.Split(string(out), "\n")
}

func GetNvmeCount() (int, error) {
	nvmeList := getNvmeDiskList()
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
