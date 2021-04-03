package fbcsnmp

import (
	log "github.com/EntropyPool/entropy-logger"
	"testing"
)

func TestCpuUsage(t *testing.T) {
	snmp := NewSnmpClient(SnmpConfig{
		target:    "172.29.100.1",
		community: "shangchi123",
		username:  "172.29.100.1",
		password:  "shangchi123",
		verbose:   true,
	})
	user, sys, idle, err := snmp.CpuUsage()
	if err != nil {
		log.Infof(log.Fields{}, "fail to get cpu usage: %v", err)
	}
	log.Infof(log.Fields{}, "cpu usages: %v | %v | %v", user, sys, idle)
}
