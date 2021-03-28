package fullnode

import (
	log "github.com/EntropyPool/entropy-logger"
	api "github.com/NpoolDevOps/fbc-devops-peer/api/lotusapi"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
	exporter "github.com/NpoolDevOps/fbc-devops-peer/exporter"
	lotusmetrics "github.com/NpoolDevOps/fbc-devops-peer/metrics/lotusmetrics"
	"github.com/prometheus/client_golang/prometheus"
)

type FullNode struct {
	*basenode.Basenode
	*lotusmetrics.LotusMetrics
}

func NewFullNode(config *basenode.BasenodeConfig, devopsClient *devops.DevopsClient) *FullNode {
	log.Infof(log.Fields{}, "create %v ndoe", config.NodeConfig.MainRole)
	fullnode := &FullNode{
		basenode.NewBasenode(config, devopsClient),
		lotusmetrics.NewLotusMetrics(),
	}

	return fullnode
}

func (n *FullNode) Describe(ch chan<- *prometheus.Desc) {
	api.ChainSyncState("127.0.0.1")
	log.Infof(log.Fields{}, "NOT IMPLEMENT FOR FULLNODE")
}

func (n *FullNode) Collect(ch chan<- prometheus.Metric) {
	log.Infof(log.Fields{}, "NOT IMPLEMENT FOR FULLNODE")
}

func (n *FullNode) CreateExporter() *exporter.Exporter {
	return exporter.NewExporter(n)
}
