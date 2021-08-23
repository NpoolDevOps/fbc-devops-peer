package devopsruntime

import (
	"testing"

	log "github.com/EntropyPool/entropy-logger"
)

func TestGetNvmeList(t *testing.T) {
	nvmeList := GetNvmeList()
	for _, nvme := range nvmeList {
		info := Info2String(nvme)
		log.Infof(log.Fields{}, "nvme: %v", info)
	}
}

func TestGetHddList(t *testing.T) {
	hddList := GetHddList()
	for _, hdd := range hddList {
		info := Info2String(hdd)
		log.Infof(log.Fields{}, "hdd: %v", info)
	}
}
