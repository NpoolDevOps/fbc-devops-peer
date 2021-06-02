package lotusmetrics

import (
	"github.com/NpoolDevOps/fbc-devops-peer/api/lotusapi"
	lotuslog "github.com/NpoolDevOps/fbc-devops-peer/loganalysis/lotuslog"
	"github.com/prometheus/client_golang/prometheus"
)

type LotusMetrics struct {
	ll *lotuslog.LotusLog

	HeightDiff         *prometheus.Desc
	BlockElapsed       *prometheus.Desc
	NetPeers           *prometheus.Desc
	LotusError         *prometheus.Desc
	SyncError          *prometheus.Desc
	ConnectionRefuseds *prometheus.Desc
	ConnectionTimeouts *prometheus.Desc
	LogFileSize        *prometheus.Desc
	LotusFileOpen      *prometheus.Desc

	host    string
	hasHost bool
	errors  int
}

func NewLotusMetrics(logfile string) *LotusMetrics {
	return &LotusMetrics{
		ll: lotuslog.NewLotusLog(logfile),
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
		ConnectionRefuseds: prometheus.NewDesc(
			"lotus_chain_net_connection_refuseds",
			"Show errors of lotus network connection refuseds",
			nil, nil,
		),
		ConnectionTimeouts: prometheus.NewDesc(
			"lotus_chain_net_connection_timeouts",
			"Show errors of lotus network connection timeouts",
			nil, nil,
		),
		LogFileSize: prometheus.NewDesc(
			"lotus_daemon_log_filesize",
			"Show daemon log filesize",
			nil, nil,
		),
		LotusFileOpen: prometheus.NewDesc(
			"Lotus_File_Opened",
			"Show Numbers File Lotus Opened",
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
	ch <- m.SyncError
	ch <- m.ConnectionRefuseds
	ch <- m.ConnectionTimeouts
	ch <- m.LogFileSize
	ch <- m.LotusFileOpen
}

func (m *LotusMetrics) Collect(ch chan<- prometheus.Metric) {
	if !m.hasHost {
		return
	}

	state, err := lotusapi.ChainSyncState(m.host)
	if err != nil {
		m.errors += 1
	}

	netPeers, err := lotusapi.ClientNetPeers(m.host)
	if err != nil {
		m.errors += 1
	}

	syncError := 0
	if state == nil || state.SyncError {
		syncError = 1
	}

	refuseds := m.ll.GetRefuseds()
	timeouts := m.ll.GetTimeouts()
	filesize := m.ll.LogFileSize()

	fileNum := lotusapi.fileWorkerOpened()
	ch <- prometheus.MustNewConstMetric(m.LotusFileOpen, prometheus.CounterValue, float64(fileNum))

	ch <- prometheus.MustNewConstMetric(m.LotusError, prometheus.CounterValue, float64(m.errors))
	if state != nil {
		ch <- prometheus.MustNewConstMetric(m.HeightDiff, prometheus.CounterValue, float64(state.HeightDiff))
		ch <- prometheus.MustNewConstMetric(m.BlockElapsed, prometheus.CounterValue, float64(state.BlockElapsed))
	}
	ch <- prometheus.MustNewConstMetric(m.NetPeers, prometheus.CounterValue, float64(netPeers))
	ch <- prometheus.MustNewConstMetric(m.SyncError, prometheus.CounterValue, float64(int(syncError)))
	ch <- prometheus.MustNewConstMetric(m.ConnectionRefuseds, prometheus.CounterValue, float64(int(refuseds)))
	ch <- prometheus.MustNewConstMetric(m.ConnectionTimeouts, prometheus.CounterValue, float64(int(timeouts)))
	ch <- prometheus.MustNewConstMetric(m.LogFileSize, prometheus.CounterValue, float64(int(filesize)))
}
