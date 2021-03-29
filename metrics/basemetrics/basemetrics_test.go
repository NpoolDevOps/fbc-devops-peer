package basemetrics

import (
	log "github.com/EntropyPool/entropy-logger"
	"testing"
)

func TestNewBaseMetrics(t *testing.T) {
	NewBaseMetrics()
	delay, lost := pingStatistic("www.baidu.com")
	log.Infof(log.Fields{}, "DELAY : %v, LOST : %v", delay, lost)
}
