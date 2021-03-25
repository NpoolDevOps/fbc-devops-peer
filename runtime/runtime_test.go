package devopsruntime

import (
	log "github.com/EntropyPool/entropy-logger"
	"testing"
)

func TestGetNvmeList(t *testing.T) {
	nvmeList := getNvmeList()
	log.Infof(log.Fields{}, "nvme list: %v", nvmeList)
}

func TestGetNvmeDesc(t *testing.T) {
	desc, _ := GetNvmeDesc()
	log.Infof(log.Fields{}, "nvme desc: %v", desc)
}

func TestGetGpuDesc(t *testing.T) {
	desc, _ := GetGpuDesc()
	log.Infof(log.Fields{}, "desc desc: %v", desc)
}

func TestGetMemoryCount(t *testing.T) {
	count, _ := GetMemoryCount()
	log.Infof(log.Fields{}, "memory count: %v", count)
}

func TestGetMemoryDesc(t *testing.T) {
	desc, _ := GetMemoryDesc()
	log.Infof(log.Fields{}, "memory desc: %v", desc)
}

func TestGetCpuCount(t *testing.T) {
	count, _ := GetCpuCount()
	log.Infof(log.Fields{}, "cpu count: %v", count)
}
