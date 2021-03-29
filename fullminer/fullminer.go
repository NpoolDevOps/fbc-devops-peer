package fullminer

import (
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
	exporter "github.com/NpoolDevOps/fbc-devops-peer/exporter"
	lotusmetrics "github.com/NpoolDevOps/fbc-devops-peer/metrics/lotusmetrics"
	"github.com/prometheus/client_golang/prometheus"
)

type FullMinerNode struct {
	*basenode.Basenode
	lotusMetrics *lotusmetrics.LotusMetrics
}

func NewFullMinerNode(config *basenode.BasenodeConfig, devopsClient *devops.DevopsClient) *FullMinerNode {
	log.Infof(log.Fields{}, "create %v node", config.NodeConfig.MainRole)
	fullminer := &FullMinerNode{
		basenode.NewBasenode(config, devopsClient),
		lotusmetrics.NewLotusMetrics(),
	}

	fullminer.SetAddrNotifier(fullminer.addressNotifier)

	return fullminer
}

func (n *FullMinerNode) addressNotifier(local, public string) {
	n.lotusMetrics.SetHost(local)
}

func (n *FullMinerNode) Describe(ch chan<- *prometheus.Desc) {
	n.lotusMetrics.Describe(ch)
	n.BaseMetrics.Describe(ch)
}

func (n *FullMinerNode) Collect(ch chan<- prometheus.Metric) {
	n.lotusMetrics.Collect(ch)
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
