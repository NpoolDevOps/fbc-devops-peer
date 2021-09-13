package lotusmetrics

import (
	"fmt"

	api "github.com/NpoolDevOps/fbc-devops-peer/api/lotusapi"
	"github.com/NpoolDevOps/fbc-devops-peer/api/systemapi"
	lotuslog "github.com/NpoolDevOps/fbc-devops-peer/loganalysis/lotuslog"
	"github.com/prometheus/client_golang/prometheus"
)

type LotusMetrics struct {
	ll *lotuslog.LotusLog

	HeightDiff          *prometheus.Desc
	BlockElapsed        *prometheus.Desc
	NetPeers            *prometheus.Desc
	LotusError          *prometheus.Desc
	SyncError           *prometheus.Desc
	ConnectionRefuseds  *prometheus.Desc
	ConnectionTimeouts  *prometheus.Desc
	LogFileSize         *prometheus.Desc
	LotusLargeDelay     *prometheus.Desc
	LotusOpenFileNumber *prometheus.Desc

	LotusRepoDirUsage   *prometheus.Desc
	LotusGatherTipsets  *prometheus.Desc
	LotusTookBlockSpent *prometheus.Desc

	host         string
	hasHost      bool
	lotusRepoDir string
	errors       int
	username     string
	networkType  string
}

func NewLotusMetrics(logfile, dir, username, networkType string) *LotusMetrics {
	return &LotusMetrics{
		ll:           lotuslog.NewLotusLog(logfile),
		lotusRepoDir: dir,
		username:     username,
		networkType:  networkType,
		HeightDiff: prometheus.NewDesc(
			"lotus_chain_height_diff",
			"Show lotus chain sync height diff",
			[]string{"networktype", "user"}, nil,
		),
		BlockElapsed: prometheus.NewDesc(
			"lotus_chain_block_elapsed",
			"Show lotus chain elapsed time of current block height",
			[]string{"networktype", "user"}, nil,
		),
		NetPeers: prometheus.NewDesc(
			"lotus_client_net_peers",
			"Show how many peers are connected by lotus client",
			[]string{"networktype", "user"}, nil,
		),
		LotusError: prometheus.NewDesc(
			"lotus_client_api_errors",
			"Show errors when request to lotus api",
			[]string{"networktype", "user"}, nil,
		),
		SyncError: prometheus.NewDesc(
			"lotus_chain_sync_error",
			"Show errors of lotus chain sync",
			[]string{"networktype", "user"}, nil,
		),
		ConnectionRefuseds: prometheus.NewDesc(
			"lotus_chain_net_connection_refuseds",
			"Show errors of lotus network connection refuseds",
			[]string{"networktype", "user"}, nil,
		),
		ConnectionTimeouts: prometheus.NewDesc(
			"lotus_chain_net_connection_timeouts",
			"Show errors of lotus network connection timeouts",
			[]string{"networktype", "user"}, nil,
		),
		LogFileSize: prometheus.NewDesc(
			"lotus_daemon_log_filesize",
			"Show daemon log filesize",
			[]string{"networktype", "user"}, nil,
		),
		LotusLargeDelay: prometheus.NewDesc(
			"lotus_large_delay",
			"show lotus large delay",
			[]string{"networktype", "user"}, nil,
		),
		LotusOpenFileNumber: prometheus.NewDesc(
			"lotus_open_file_number",
			"show lotus open file number",
			[]string{"networktype", "user"}, nil,
		),
		LotusRepoDirUsage: prometheus.NewDesc(
			"lotus_repo_dir_usage",
			"show lotus repo dir usage",
			[]string{"repodir", "totalcap", "networktype", "user"}, nil,
		),
		LotusGatherTipsets: prometheus.NewDesc(
			"lotus_gather_tipsets",
			"show lotus gather tipsets number",
			[]string{"networktype", "user"}, nil,
		),
		LotusTookBlockSpent: prometheus.NewDesc(
			"lotus_took_blocks_spent",
			"show lotus took blocks spent",
			[]string{"networktype", "user"}, nil,
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
	ch <- m.LotusLargeDelay
	ch <- m.LotusOpenFileNumber
	ch <- m.LotusRepoDirUsage
	ch <- m.LotusGatherTipsets
	ch <- m.LotusTookBlockSpent
}

func (m *LotusMetrics) Collect(ch chan<- prometheus.Metric) {
	if !m.hasHost {
		return
	}

	username := m.username
	networkType := m.networkType

	state, err := api.ChainSyncState(m.host)
	if err != nil {
		m.errors += 1
	}

	netPeers, err := api.ClientNetPeers(m.host)
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
	largeDelay := m.ll.GetLargeDelay()

	ch <- prometheus.MustNewConstMetric(m.LotusError, prometheus.CounterValue, float64(m.errors), networkType, username)
	if state != nil {
		ch <- prometheus.MustNewConstMetric(m.HeightDiff, prometheus.CounterValue, float64(state.HeightDiff), networkType, username)
		ch <- prometheus.MustNewConstMetric(m.BlockElapsed, prometheus.CounterValue, float64(state.BlockElapsed.Milliseconds()), networkType, username)
	}
	ch <- prometheus.MustNewConstMetric(m.NetPeers, prometheus.CounterValue, float64(netPeers), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.SyncError, prometheus.CounterValue, float64(int(syncError)), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.ConnectionRefuseds, prometheus.CounterValue, float64(int(refuseds)), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.ConnectionTimeouts, prometheus.CounterValue, float64(int(timeouts)), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.LogFileSize, prometheus.CounterValue, float64(int(filesize)), networkType, username)
	ch <- prometheus.MustNewConstMetric(m.LotusLargeDelay, prometheus.CounterValue, largeDelay, networkType, username)

	lotusOpenFileNumber, err := systemapi.GetProcessOpenFileNumber("lotus")
	if err != nil {
		ch <- prometheus.MustNewConstMetric(m.LotusOpenFileNumber, prometheus.CounterValue, 0, networkType, username)
	}
	ch <- prometheus.MustNewConstMetric(m.LotusOpenFileNumber, prometheus.CounterValue, float64(lotusOpenFileNumber), networkType, username)

	dirStatus, dirPath := getFullnodeRepoDirUsage(m.lotusRepoDir)
	ch <- prometheus.MustNewConstMetric(m.LotusRepoDirUsage, prometheus.CounterValue, dirStatus.Used, fmt.Sprintf("%v", dirPath), fmt.Sprintf("%v", dirStatus.All), networkType, username)

	tipset := m.ll.GetGatherTipsets()
	spent := m.ll.GetTookBlocksSpent()
	ch <- prometheus.MustNewConstMetric(m.LotusGatherTipsets, prometheus.CounterValue, tipset, networkType, username)
	ch <- prometheus.MustNewConstMetric(m.LotusTookBlockSpent, prometheus.CounterValue, spent, networkType, username)
}

func getFullnodeRepoDirUsage(dir string) (systemapi.DiskStatus, string) {
	return systemapi.DiskUsage(dir), dir
}
