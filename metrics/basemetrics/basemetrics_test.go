package basemetrics

import (
	"fmt"
	"testing"

	log "github.com/EntropyPool/entropy-logger"
)

func TestNewBaseMetrics(t *testing.T) {
	// NewBaseMetrics("entropytest")
	delay, lost := pingStatistic("www.baidu.com")
	log.Infof(log.Fields{}, "DELAY : %v, LOST : %v", delay, lost)
}

func TestNtpTimeDiff(t *testing.T) {
	diff, err := getNtpDiff()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(diff)
}
