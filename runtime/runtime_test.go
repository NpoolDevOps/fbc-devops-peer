package devopsruntime

import (
	log "github.com/EntropyPool/entropy-logger"
	"testing"
)

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

func TestGetCpuDesc(t *testing.T) {
	desc, _ := GetCpuDesc()
	log.Infof(log.Fields{}, "cpu desc: %v", desc)
}

func TestGetHddCount(t *testing.T) {
	count, _ := GetHddCount()
	log.Infof(log.Fields{}, "hdd count: %v", count)
}

func TestGetHddDesc(t *testing.T) {
	desc, _ := GetHddDesc()
	log.Infof(log.Fields{}, "hdd desc: %v", desc)
}

func TestGetEthernetCount(t *testing.T) {
	count, err := GetEthernetCount()
	log.Infof(log.Fields{}, "ethernet count: %v | %v", count, err)
}

func TestGetEthernetDesc(t *testing.T) {
	desc, err := GetEthernetDesc()
	log.Infof(log.Fields{}, "ethernet desc: %v | %v", desc, err)
}
