package minermetrics

import (
	"github.com/NpoolDevOps/fbc-devops-peer/loganalysis/minerlog"
	"github.com/prometheus/client_golang/prometheus"
)

type MinerMetrics struct {
	ml          *minerlog.MinerLog
	ForkBlocks  *prometheus.Desc
	BlockTookMs *prometheus.Desc

	errors       int
	host         string
	hasHost      bool
	fullnodeHost string
}

func NewMinerMetrics(logfile string) *MinerMetrics {
	mm := &MinerMetrics{
		ml: minerlog.NewMinerLog(logfile),
		ForkBlocks: prometheus.NewDesc(
			"miner_fork_blocks",
			"Show miner fork blocks",
			nil, nil,
		),
		BlockTookMs: prometheus.NewDesc(
			"miner_block_took_ms",
			"Show miner block took ms",
			nil, nil,
		),
	}
	return mm
}

func (m *MinerMetrics) SetHost(host string) {
	m.host = host
}

func (m *MinerMetrics) SetFullnodeHost(host string) {
	m.ml.SetFullnodeHost(host)
}

func (m *MinerMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.ForkBlocks
	ch <- m.BlockTookMs
}

func (m *MinerMetrics) Collect(ch chan<- prometheus.Metric) {
	tooks := m.ml.GetBlockTooks()
	forkBlocks := m.ml.GetForkBlocks()

	ch <- prometheus.MustNewConstMetric(m.ForkBlocks, prometheus.CounterValue, float64(forkBlocks))
	for _, took := range tooks {
		ch <- prometheus.MustNewConstMetric(m.BlockTookMs, prometheus.CounterValue, float64(took))
	}
}
