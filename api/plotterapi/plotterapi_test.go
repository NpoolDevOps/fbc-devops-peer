package plotterapi

import (
	"testing"

	log "github.com/EntropyPool/entropy-logger"
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
