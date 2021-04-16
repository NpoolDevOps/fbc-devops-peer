package minermetrics

import (
	"github.com/NpoolDevOps/fbc-devops-peer/loganalysis/minerlog"
	"github.com/prometheus/client_golang/prometheus"
)

type MinerMetrics struct {
	ml             *minerlog.MinerLog
	ForkBlocks     *prometheus.Desc
	PastBlocks     *prometheus.Desc
	FailedBlocks   *prometheus.Desc
	BlockTookAvgMs *prometheus.Desc
	BlockTookMaxMs *prometheus.Desc
	BlockTookMinMs *prometheus.Desc
	Blocks         *prometheus.Desc

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
		PastBlocks: prometheus.NewDesc(
			"miner_block_in_past",
			"Show miner block in past",
			nil, nil,
		),
		FailedBlocks: prometheus.NewDesc(
			"miner_block_failed",
			"Show miner block failed",
			nil, nil,
		),
		BlockTookAvgMs: prometheus.NewDesc(
			"miner_block_took_average_ms",
			"Show miner block took average ms",
			nil, nil,
		),
		BlockTookMaxMs: prometheus.NewDesc(
			"miner_block_took_max_ms",
			"Show miner block took max ms",
			nil, nil,
		),
		BlockTookMinMs: prometheus.NewDesc(
			"miner_block_took_min_ms",
			"Show miner block took min ms",
			nil, nil,
		),
		Blocks: prometheus.NewDesc(
			"miner_block_produced",
			"Show miner block produced",
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
	ch <- m.PastBlocks
	ch <- m.FailedBlocks
	ch <- m.BlockTookAvgMs
	ch <- m.BlockTookMaxMs
	ch <- m.BlockTookMinMs
	ch <- m.Blocks
}

func (m *MinerMetrics) Collect(ch chan<- prometheus.Metric) {
	tooks := m.ml.GetBlockTooks()
	forkBlocks := m.ml.GetForkBlocks()
	pastBlocks := m.ml.GetPastBlocks()
	failedBlocks := m.ml.GetFailedBlocks()

	avgMs := uint64(0)
	maxMs := uint64(0)
	minMs := uint64(0)

	for _, took := range tooks {
		avgMs += took
		if maxMs < took {
			maxMs = took
		}
		if minMs == 0 || took < minMs {
			minMs = took
		}
	}
	avgMs = avgMs / uint64(len(tooks))

	ch <- prometheus.MustNewConstMetric(m.ForkBlocks, prometheus.CounterValue, float64(forkBlocks))
	ch <- prometheus.MustNewConstMetric(m.PastBlocks, prometheus.CounterValue, float64(pastBlocks))
	ch <- prometheus.MustNewConstMetric(m.FailedBlocks, prometheus.CounterValue, float64(failedBlocks))
	ch <- prometheus.MustNewConstMetric(m.BlockTookAvgMs, prometheus.CounterValue, float64(avgMs))
	ch <- prometheus.MustNewConstMetric(m.BlockTookMaxMs, prometheus.CounterValue, float64(maxMs))
	ch <- prometheus.MustNewConstMetric(m.BlockTookMinMs, prometheus.CounterValue, float64(minMs))
	ch <- prometheus.MustNewConstMetric(m.Blocks, prometheus.CounterValue, float64(len(tooks)))
}
