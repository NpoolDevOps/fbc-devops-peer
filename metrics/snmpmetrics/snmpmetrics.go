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
	snmpClient             *snmp.SnmpClient
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
		SnmpError: prometheus.NewDesc(
			"switcher_snmp_error",
			"Switcher snmp error",
			nil, nil,
		),
		snmpClient: snmp.NewSnmpClient(config),
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
}
