package fullminer

import (
	log "github.com/EntropyPool/entropy-logger"
	api "github.com/NpoolDevOps/fbc-devops-peer/api/lotusapi"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
	exporter "github.com/NpoolDevOps/fbc-devops-peer/exporter"
	lotusmetrics "github.com/NpoolDevOps/fbc-devops-peer/metrics/lotusmetrics"
	"github.com/prometheus/client_golang/prometheus"
)

type FullMinerNode struct {
	*basenode.Basenode
	*lotusmetrics.LotusMetrics
}

func NewFullMinerNode(config *basenode.BasenodeConfig, devopsClient *devops.DevopsClient) *FullMinerNode {
	log.Infof(log.Fields{}, "create %v ndoe", config.NodeConfig.MainRole)
	fullminer := &FullMinerNode{
		basenode.NewBasenode(config, devopsClient),
		lotusmetrics.NewLotusMetrics(),
	}
	return fullminer
}

func (n *FullMinerNode) Describe(ch chan<- *prometheus.Desc) {
	log.Infof(log.Fields{}, "NOT IMPLEMENT FOR FULLMINERNODE")
}

func (n *FullMinerNode) Collect(ch chan<- prometheus.Metric) {
	addr, err := n.MyLocalAddr()
	if err != nil {
		return
	}

	state, err := api.ChainSyncState(addr)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to check chain sync status %v: %v", addr, err)
	}
	log.Infof(log.Fields{}, "state --- %v", state)

	log.Infof(log.Fields{}, "NOT IMPLEMENT FOR FULLMINERNODE")
}

func (n *FullMinerNode) CreateExporter() *exporter.Exporter {
	return exporter.NewExporter(n)
}

func (n *FullMinerNode) Banner() {
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
	log.Infof(log.Fields{}, "   FUlllllLLLLMMMMMMMIIIiNEEeer    ")
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
}
