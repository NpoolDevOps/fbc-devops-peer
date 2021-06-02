package fullminer

import (
	log "github.com/EntropyPool/entropy-logger"
	lotusapi "github.com/NpoolDevOps/fbc-devops-peer/api/lotusapi"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
	exporter "github.com/NpoolDevOps/fbc-devops-peer/exporter"
	lotusmetrics "github.com/NpoolDevOps/fbc-devops-peer/metrics/lotusmetrics"
	minermetrics "github.com/NpoolDevOps/fbc-devops-peer/metrics/minermetrics"
	types "github.com/NpoolDevOps/fbc-devops-peer/types"
	version "github.com/NpoolDevOps/fbc-devops-peer/version"
	"github.com/prometheus/client_golang/prometheus"
)

type FullMinerNode struct {
	*basenode.Basenode
	lotusMetrics *lotusmetrics.LotusMetrics
	minerMetrics *minermetrics.MinerMetrics
}

func NewFullMinerNode(config *basenode.BasenodeConfig, devopsClient *devops.DevopsClient) *FullMinerNode {
	log.Infof(log.Fields{}, "create %v node", config.NodeConfig.MainRole)
	fullminer := &FullMinerNode{
		basenode.NewBasenode(config, devopsClient),
		nil, nil,
	}

	logfile, _ := fullminer.GetLogFileByRole(types.FullNode)
	fullminer.lotusMetrics = lotusmetrics.NewLotusMetrics(logfile)
	logfile, _ = fullminer.GetLogFileByRole(types.MinerNode)
	fullminer.minerMetrics = minermetrics.NewMinerMetrics(logfile)

	fullminer.SetAddrNotifier(fullminer.addressNotifier)
	fullminer.WatchVersions(fullminer.getVersions)

	return fullminer
}

func (n *FullMinerNode) addressNotifier(local, public string) {
	n.lotusMetrics.SetHost(local)
	n.minerMetrics.SetHost(local)
	n.minerMetrics.SetFullnodeHost(local)
}

func (n *FullMinerNode) getVersions(host string) []version.Version {
	vers := []version.Version{}

	ver, err := lotusapi.ClientVersion(host)
	if err == nil {
		vers = append(vers, ver)
	}

	return vers
}

func (n *FullMinerNode) Describe(ch chan<- *prometheus.Desc) {
	n.lotusMetrics.Describe(ch)
	n.minerMetrics.Describe(ch)
	n.BaseMetrics.Describe(ch)
}

func (n *FullMinerNode) Collect(ch chan<- prometheus.Metric) {
	n.lotusMetrics.Collect(ch)
	n.minerMetrics.Collect(ch)
	n.BaseMetrics.Collect(ch)
}

func (n *FullMinerNode) CreateExporter() *exporter.Exporter {
	return exporter.NewExporter(n)
}

func (n *FullMinerNode) Banner() {
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
	log.Infof(log.Fields{}, "   FUlllllLLLLMMMMMMMIIIiNEEeer    ")
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
}
