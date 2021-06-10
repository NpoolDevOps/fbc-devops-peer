package plotterapi

import (
	"testing"

	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/api/chiaapi"
)

func TestGetPlotterProcessNum(t *testing.T) {
	num, err := GetPlotterProcessNum()
	if err != nil {
		log.Infof(log.Fields{}, "err is: %v", err)
	}
	log.Infof(log.Fields{}, "process num is:%v", num)
}

func TestGetPlotterTime(t *testing.T) {
	time, count, err := GetPlotterTime()
	if err != nil {
		log.Infof(log.Fields{}, "err is: %v", err)
	}
	log.Infof(log.Fields{}, "process time is:%v, count is:%v", time, count)
}

func TestGetStatus(t *testing.T) {
	status1, err := chiaapi.GetStatus("ProofOfSpace create")
	if err != nil {
		log.Infof(log.Fields{}, "err is: %v", err)
	}
	log.Infof(log.Fields{}, "plotter status is:%v", status1)

	status2, err := chiaapi.GetStatus("chia-storage-proxy")
	if err != nil {
		log.Infof(log.Fields{}, "err is: %v", err)
	}
	log.Infof(log.Fields{}, "chia-storage-proxy status is:%v", status2)
}
