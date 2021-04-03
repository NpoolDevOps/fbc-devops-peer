package snmpmetrics

import (
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
	NetworkBandwidth       *prometheus.Desc
	NetworkConfigBandwidth *prometheus.Desc
	NetworkRecvBytes       *prometheus.Desc
	NetworkSendBytes       *prometheus.Desc
	snmpClient             *snmp.SnmpClient
	label                  string
}

func NewSnmpMetrics(config *snmp.SnmpConfig) *SnmpMetrics {
	return &SnmpMetrics{
		CpuUserPercent: prometheus.NewDesc(
			"switcher_cpu_user_percent",
			"Switcher cpu user percent",
			nil, nil,
		),
		CpuSysPercent: prometheus.NewDesc(
			"switcher_cpu_sys_percent",
			"Switcher cpu sys percent",
			nil, nil,
		),
		CpuIdlePercent: prometheus.NewDesc(
			"switcher_cpu_idle_percent",
			"Switcher cpu idle percent",
			nil, nil,
		),
		MemTotalReal: prometheus.NewDesc(
			"switcher_mem_total_real",
			"Switcher mem total real",
			nil, nil,
		),
		MemUsedReal: prometheus.NewDesc(
			"switcher_mem_used_real",
			"Switcher mem used real",
			nil, nil,
		),
		NetworkBandwidth: prometheus.NewDesc(
			"switcher_network_bandwidth",
			"Switcher network bandwidth",
			nil, nil,
		),
		NetworkConfigBandwidth: prometheus.NewDesc(
			"switcher_network_config_bandwidth",
			"Switcher network config bandwidth",
			nil, nil,
		),
		NetworkRecvBytes: prometheus.NewDesc(
			"switcher_network_recv_bytes",
			"Switcher network recv bytes",
			nil, nil,
		),
		NetworkSendBytes: prometheus.NewDesc(
			"switcher_network_send_bytes",
			"Switcher network send bytes",
			nil, nil,
		),
		SnmpError: prometheus.NewDesc(
			"switcher_snmp_error",
			"Switcher snmp error",
			nil, nil,
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
	ch <- m.NetworkBandwidth
	ch <- m.NetworkConfigBandwidth
	ch <- m.SnmpError
}

func (m *SnmpMetrics) Collect(ch chan<- prometheus.Metric) {
	snmpError := 0

	cpuUser, cpuSys, cpuIdle, err := m.snmpClient.CpuUsage()
	if err != nil {
		snmpError += 1
	}

	bw, configBw, err := m.snmpClient.NetworkBandwidth()
	if err != nil {
		snmpError += 1
	}

	recvBytes, sendBytes, err := m.snmpClient.NetworkBytes()
	if err != nil {
		snmpError += 1
	}

	ch <- prometheus.MustNewConstMetric(m.CpuUserPercent, prometheus.CounterValue, float64(cpuUser), m.label)
	ch <- prometheus.MustNewConstMetric(m.CpuIdlePercent, prometheus.CounterValue, float64(cpuSys), m.label)
	ch <- prometheus.MustNewConstMetric(m.CpuSysPercent, prometheus.CounterValue, float64(cpuIdle), m.label)
	ch <- prometheus.MustNewConstMetric(m.MemTotalReal, prometheus.CounterValue, float64(0), m.label)
	ch <- prometheus.MustNewConstMetric(m.MemUsedReal, prometheus.CounterValue, float64(0), m.label)
	ch <- prometheus.MustNewConstMetric(m.NetworkBandwidth, prometheus.CounterValue, float64(bw), m.label)
	ch <- prometheus.MustNewConstMetric(m.NetworkConfigBandwidth, prometheus.CounterValue, float64(configBw), m.label)
	ch <- prometheus.MustNewConstMetric(m.NetworkRecvBytes, prometheus.CounterValue, float64(recvBytes), m.label)
	ch <- prometheus.MustNewConstMetric(m.NetworkSendBytes, prometheus.CounterValue, float64(sendBytes), m.label)
	ch <- prometheus.MustNewConstMetric(m.SnmpError, prometheus.CounterValue, float64(snmpError), m.label)
}
