package fbcsnmp

import (
	log "github.com/EntropyPool/entropy-logger"
	"testing"
)

func TestCpuUsage(t *testing.T) {
	snmp := NewSnmpClient(&SnmpConfig{
		Target:    "172.29.100.1",
		Community: "shangchi123",
		Username:  "172.29.100.1",
		Password:  "shangchi123",
		verbose:   true,
	})
	user, sys, idle, err := snmp.CpuUsage()
	if err != nil {
		log.Infof(log.Fields{}, "fail to get cpu usage: %v", err)
	}
	log.Infof(log.Fields{}, "cpu usages: %v | %v | %v", user, sys, idle)
}

func TestNetwork(t *testing.T) {
	snmp := NewSnmpClient(&SnmpConfig{
		Target:          "172.29.100.1",
		Community:       "shangchi123",
		Username:        "172.29.100.1",
		Password:        "shangchi123",
		verbose:         true,
		ConfigBandwidth: 500 * 1024 * 1024,
	})
	bw, configBw, err := snmp.NetworkBandwidth()
	if err != nil {
		log.Infof(log.Fields{}, "fail to get network bandwidth: %v", err)
	}
	log.Infof(log.Fields{}, "network bandwidth: %v | %v", bw, configBw)

	recvBytes, sendBytes, err := snmp.NetworkBytes()
	if err != nil {
		log.Infof(log.Fields{}, "fail to get network bytes: %v", err)
	}
	log.Infof(log.Fields{}, "network bytes: %v | %v", recvBytes, sendBytes)
}
