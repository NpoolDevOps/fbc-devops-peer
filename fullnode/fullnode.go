package fullnode

import (
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/api/lotusapi"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
	exporter "github.com/NpoolDevOps/fbc-devops-peer/exporter"
	lotusmetrics "github.com/NpoolDevOps/fbc-devops-peer/metrics/lotusmetrics"
	"github.com/NpoolDevOps/fbc-devops-peer/version"
	"github.com/prometheus/client_golang/prometheus"
)

type FullNode struct {
	*basenode.Basenode
	lotusMetrics *lotusmetrics.LotusMetrics
}

func NewFullNode(config *basenode.BasenodeConfig, devopsClient *devops.DevopsClient) *FullNode {
	log.Infof(log.Fields{}, "create %v node", config.NodeConfig.MainRole)
	fullnode := &FullNode{
		basenode.NewBasenode(config, devopsClient),
		nil,
	}

	logfile, _ := fullnode.GetLogFile()
	fullnode.lotusMetrics = lotusmetrics.NewLotusMetrics(logfile)

	fullnode.SetAddrNotifier(func(local, public string) {
		fullnode.lotusMetrics.SetHost(local)
	})

	localAddr := fullnode.GetFullnodeLocalAddr()
	fullnode.WatchVersions(localAddr, nil, fullnode.getVersions)

	return fullnode
}

func (n *FullNode) getVersions(host string) []version.Version {
	vers := []version.Version{}

	ver, err := lotusapi.ClientVersion(host)
	if err == nil {
		vers = append(vers, ver)
	}

	return vers
}

func (n *FullNode) Describe(ch chan<- *prometheus.Desc) {
	n.lotusMetrics.Describe(ch)
	n.BaseMetrics.Describe(ch)
}

func (n *FullNode) Collect(ch chan<- prometheus.Metric) {
	n.lotusMetrics.Collect(ch)
	n.BaseMetrics.Collect(ch)
}

func (n *FullNode) CreateExporter() *exporter.Exporter {
	return exporter.NewExporter(n)
}
