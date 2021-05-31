package basemetrics

import (
	"testing"

	log "github.com/EntropyPool/entropy-logger"
)

func TestNewBaseMetrics(t *testing.T) {
	NewBaseMetrics()
	delay, lost := pingStatistic("www.baidu.com")
	log.Infof(log.Fields{}, "DELAY : %v, LOST : %v", delay, lost)
}
