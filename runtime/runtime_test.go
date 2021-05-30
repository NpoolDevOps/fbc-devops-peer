package runtime

import (
	"testing"

	log "github.com/EntropyPool/entropy-logger"
)

func TestGetNvmeList(t *testing.T) {
	nvmeList := GetNvmeList()
	log.Infof(log.Fields{}, "nvme list: %v", nvmeList)
}
func TestGetNvmeCount(t *testing.T) {
	count, _ := GetNvmeCount()
	log.Infof(log.Fields{}, "nvme count: %v", count)
}
func TestGetNvmeDesc(t *testing.T) {
	desc, _ := GetNvmeDesc()
	log.Infof(log.Fields{}, "nvme desc: %v", desc)
}

func TestGetHddList(t *testing.T) {
	hddList := GetHddList()
	log.Infof(log.Fields{}, "hdd list : %v", hddList)
}
func TestGetHddCount(t *testing.T) {
	count, _ := GetHddCount()
	log.Infof(log.Fields{}, "hdd count: %v", count)
}
func TestGetHddDesc(t *testing.T) {
	desc, _ := GetHddDesc()
	log.Infof(log.Fields{}, "hdd desc: %v", desc)
}

func TestGetSsdsList(t *testing.T) {
	ssdList := GetSsdsList()
	log.Infof(log.Fields{}, "ssd list : %v", ssdList)
}
func TestGetSsdsCount(t *testing.T) {
	ssdCount, _ := GetSsdsCount()
	log.Infof(log.Fields{}, "ssd count : %v", ssdCount)
}
func TestGetSsdsDesc(t *testing.T) {
	ssdDesc, _ := GetSsdsDesc()
	log.Infof(log.Fields{}, "ssd desc : %v", ssdDesc)
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

func TestGetEthernetCount(t *testing.T) {
	count, err := GetEthernetCount()
	log.Infof(log.Fields{}, "ethernet count: %v | %v", count, err)
}

func TestGetEthernetDesc(t *testing.T) {
	desc, err := GetEthernetDesc()
	log.Infof(log.Fields{}, "ethernet desc: %v | %v", desc, err)
}
