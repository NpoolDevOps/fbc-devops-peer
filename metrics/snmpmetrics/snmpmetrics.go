package snmpmetrics

import (
	log "github.com/EntropyPool/entropy-logger"
	snmp "github.com/NpoolDevOps/fbc-devops-peer/snmp"
	"github.com/prometheus/client_golang/prometheus"
)

type SnmpMetrics struct {
	CpuUserPercent         *prometheus.Desc
	CpuSysPercent          *prometheus.Desc
	CpuIdlePercent         *prometheus.Desc
	CpuProcessorLoad       *prometheus.Desc
	MemTotalReal           *prometheus.Desc
	MemUsedReal            *prometheus.Desc
	SnmpError              *prometheus.Desc
	NetworkInBandwidth     *prometheus.Desc
	NetworkOutBandwidth    *prometheus.Desc
	NetworkConfigBandwidth *prometheus.Desc
	NetworkRecvBytes       *prometheus.Desc
	NetworkSendBytes       *prometheus.Desc
	OutDiscards            *prometheus.Desc
	OutErrors              *prometheus.Desc
	MemorySize             *prometheus.Desc
	snmpClient             *snmp.SnmpClient
	label                  string
}

func NewSnmpMetrics(config *snmp.SnmpConfig) *SnmpMetrics {
	return &SnmpMetrics{
		CpuUserPercent: prometheus.NewDesc(
			"switcher_cpu_user_percent",
			"Switcher cpu user percent",
			[]string{"location"}, nil,
		),
		CpuSysPercent: prometheus.NewDesc(
			"switcher_cpu_sys_percent",
			"Switcher cpu sys percent",
			[]string{"localtion"}, nil,
		),
		CpuIdlePercent: prometheus.NewDesc(
			"switcher_cpu_idle_percent",
			"Switcher cpu idle percent",
			[]string{"location"}, nil,
		),
		MemTotalReal: prometheus.NewDesc(
			"switcher_mem_total_real",
			"Switcher mem total real",
			[]string{"location"}, nil,
		),
		MemUsedReal: prometheus.NewDesc(
			"switcher_mem_used_real",
			"Switcher mem used real",
			[]string{"location"}, nil,
		),
		NetworkInBandwidth: prometheus.NewDesc(
			"switcher_network_in_bandwidth",
			"Switcher network in bandwidth",
			[]string{"location"}, nil,
		),
		NetworkOutBandwidth: prometheus.NewDesc(
			"switcher_network_out_bandwidth",
			"Switcher network out bandwidth",
			[]string{"location"}, nil,
		),
		NetworkConfigBandwidth: prometheus.NewDesc(
			"switcher_network_config_bandwidth",
			"Switcher network config bandwidth",
			[]string{"location"}, nil,
		),
		NetworkRecvBytes: prometheus.NewDesc(
			"switcher_network_recv_bytes",
			"Switcher network recv bytes",
			[]string{"location"}, nil,
		),
		NetworkSendBytes: prometheus.NewDesc(
			"switcher_network_send_bytes",
			"Switcher network send bytes",
			[]string{"location"}, nil,
		),
		OutDiscards: prometheus.NewDesc(
			"switcher_out_discards",
			"Switcher out discards",
			[]string{"location"}, nil,
		),
		OutErrors: prometheus.NewDesc(
			"switcher_out_errors",
			"Switcher out errors",
			[]string{"location"}, nil,
		),
		MemorySize: prometheus.NewDesc(
			"switcher_memory_size",
			"Switcher memory size",
			[]string{"location"}, nil,
		),
		SnmpError: prometheus.NewDesc(
			"switcher_snmp_error",
			"Switcher snmp error",
			[]string{"location"}, nil,
		),
		snmpClient: snmp.NewSnmpClient(config),
		label:      config.Label,
	}
}

func (m *SnmpMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.CpuUserPercent
	ch <- m.CpuIdlePercent
	ch <- m.CpuSysPercent
	ch <- m.MemTotalReal
	ch <- m.MemUsedReal
	ch <- m.NetworkInBandwidth
	ch <- m.NetworkOutBandwidth
	ch <- m.NetworkConfigBandwidth
	ch <- m.OutDiscards
	ch <- m.OutErrors
	ch <- m.MemorySize
	ch <- m.SnmpError
}

func (m *SnmpMetrics) Collect(ch chan<- prometheus.Metric) {
	snmpError := 0

	cpuUser, cpuSys, cpuIdle, err := m.snmpClient.CpuUsage()
	if err != nil {
		log.Errorf(log.Fields{}, "fail to get cpu usage: %v", err)
		snmpError += 1
	}

	inbw, outbw, configBw, err := m.snmpClient.NetworkBandwidth()
	if err != nil {
		log.Errorf(log.Fields{}, "fail to get network bandwidth: %v", err)
		snmpError += 1
	}

	recvBytes, sendBytes, err := m.snmpClient.NetworkBytes()
	if err != nil {
		log.Errorf(log.Fields{}, "fail to get network bytes: %v", err)
		snmpError += 1
	}

	outDiscards, err := m.snmpClient.OutDiscards()
	if err != nil {
		log.Errorf(log.Fields{}, "fail to get out discards: %v", err)
		snmpError += 1
	}

	outErrors, err := m.snmpClient.OutErrors()
	if err != nil {
		log.Errorf(log.Fields{}, "fail to get out errors: %v", err)
		snmpError += 1
	}

	memorySize, err := m.snmpClient.MemorySize()
	if err != nil {
		log.Errorf(log.Fields{}, "fail to get memory size: %v", err)
		snmpError += 1
	}

	ch <- prometheus.MustNewConstMetric(m.CpuUserPercent, prometheus.CounterValue, float64(cpuUser), m.label)
	ch <- prometheus.MustNewConstMetric(m.CpuIdlePercent, prometheus.CounterValue, float64(cpuSys), m.label)
	ch <- prometheus.MustNewConstMetric(m.CpuSysPercent, prometheus.CounterValue, float64(cpuIdle), m.label)
	ch <- prometheus.MustNewConstMetric(m.MemTotalReal, prometheus.CounterValue, float64(0), m.label)
	ch <- prometheus.MustNewConstMetric(m.MemUsedReal, prometheus.CounterValue, float64(0), m.label)
	ch <- prometheus.MustNewConstMetric(m.NetworkInBandwidth, prometheus.CounterValue, float64(inbw), m.label)
	ch <- prometheus.MustNewConstMetric(m.NetworkOutBandwidth, prometheus.CounterValue, float64(outbw), m.label)
	ch <- prometheus.MustNewConstMetric(m.NetworkConfigBandwidth, prometheus.CounterValue, float64(configBw), m.label)
	ch <- prometheus.MustNewConstMetric(m.NetworkRecvBytes, prometheus.CounterValue, float64(recvBytes), m.label)
	ch <- prometheus.MustNewConstMetric(m.NetworkSendBytes, prometheus.CounterValue, float64(sendBytes), m.label)
	ch <- prometheus.MustNewConstMetric(m.OutDiscards, prometheus.CounterValue, float64(outDiscards), m.label)
	ch <- prometheus.MustNewConstMetric(m.OutErrors, prometheus.CounterValue, float64(outErrors), m.label)
	ch <- prometheus.MustNewConstMetric(m.MemorySize, prometheus.CounterValue, float64(memorySize), m.label)
	ch <- prometheus.MustNewConstMetric(m.SnmpError, prometheus.CounterValue, float64(snmpError), m.label)
}
