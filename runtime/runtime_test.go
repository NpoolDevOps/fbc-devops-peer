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
