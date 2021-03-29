package lotusmetrics

import (
	api "github.com/NpoolDevOps/fbc-devops-peer/api/lotusapi"
	"github.com/prometheus/client_golang/prometheus"
)

type LotusMetrics struct {
	HeightDiff   *prometheus.Desc
	BlockElapsed *prometheus.Desc
	NetPeers     *prometheus.Desc
	LotusError   *prometheus.Desc
	SyncError    *prometheus.Desc

	host    string
	hasHost bool
	errors  int
}

func NewLotusMetrics() *LotusMetrics {
	return &LotusMetrics{
		HeightDiff: prometheus.NewDesc(
			"lotus_chain_height_diff",
			"Show lotus chain sync height diff",
			nil, nil,
		),
		BlockElapsed: prometheus.NewDesc(
			"lotus_chain_block_elapsed",
			"Show lotus chain elapsed time of current block height",
			nil, nil,
		),
		NetPeers: prometheus.NewDesc(
			"lotus_client_net_peers",
			"Show how many peers are connected by lotus client",
			nil, nil,
		),
		LotusError: prometheus.NewDesc(
			"lotus_client_api_errors",
			"Show errors when request to lotus api",
			nil, nil,
		),
		SyncError: prometheus.NewDesc(
			"lotus_chain_sync_error",
			"Show errors of lotus chain sync",
			nil, nil,
		),
	}
}

func (m *LotusMetrics) SetHost(host string) {
	m.host = host
	m.hasHost = true
}

func (m *LotusMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.HeightDiff
	ch <- m.BlockElapsed
	ch <- m.NetPeers
	ch <- m.LotusError
}

func (m *LotusMetrics) Collect(ch chan<- prometheus.Metric) {
	if !m.hasHost {
		return
	}

	state, err := api.ChainSyncState(m.host)
	if err != nil {
		m.errors += 1
	}

	netPeers, err := api.ClientNetPeers(m.host)
	if err != nil {
		m.errors += 1
	}

	syncError := 0
	if state.SyncError {
		syncError = 1
	}

	ch <- prometheus.MustNewConstMetric(m.LotusError, prometheus.CounterValue, float64(m.errors))
	ch <- prometheus.MustNewConstMetric(m.HeightDiff, prometheus.CounterValue, float64(state.HeightDiff))
	ch <- prometheus.MustNewConstMetric(m.BlockElapsed, prometheus.CounterValue, float64(state.BlockElapsed))
	ch <- prometheus.MustNewConstMetric(m.NetPeers, prometheus.CounterValue, float64(netPeers))
	ch <- prometheus.MustNewConstMetric(m.SyncError, prometheus.CounterValue, float64(int(syncError)))
}
