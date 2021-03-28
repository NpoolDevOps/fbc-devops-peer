package fullminer

import (
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
	exporter "github.com/NpoolDevOps/fbc-devops-peer/exporter"
	"github.com/prometheus/client_golang/prometheus"
)

type FullMinerNode struct {
	*basenode.Basenode
}

func NewFullMinerNode(config *basenode.BasenodeConfig, devopsClient *devops.DevopsClient) *FullMinerNode {
	log.Infof(log.Fields{}, "create %v ndoe", config.NodeConfig.MainRole)
	fullminer := &FullMinerNode{
		basenode.NewBasenode(config, devopsClient),
	}
	return fullminer
}

func (n *FullMinerNode) Describe(ch chan<- *prometheus.Desc) {
	log.Infof(log.Fields{}, "NOT IMPLEMENT FOR FULLMINERNODE")
}

func (n *FullMinerNode) Collect(ch chan<- prometheus.Metric) {
	log.Infof(log.Fields{}, "NOT IMPLEMENT FOR FULLMINERNODE")
}

func (n *FullMinerNode) CreateExporter() *exporter.Exporter {
	return exporter.NewExporter(n)
}
