package miner

import (
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
	exporter "github.com/NpoolDevOps/fbc-devops-peer/exporter"
	"github.com/prometheus/client_golang/prometheus"
)

type MinerNode struct {
	*basenode.Basenode
}

func NewMinerNode(config *basenode.BasenodeConfig, devopsClient *devops.DevopsClient) *MinerNode {
	log.Infof(log.Fields{}, "create %v ndoe", config.NodeConfig.MainRole)
	miner := &MinerNode{
		basenode.NewBasenode(config, devopsClient),
	}
	return miner
}

func (n *MinerNode) Describe(ch chan<- *prometheus.Desc) {
	log.Infof(log.Fields{}, "NOT IMPLEMENT FOR MINERNODE")
}

func (n *MinerNode) Collect(ch chan<- prometheus.Metric) {
	log.Infof(log.Fields{}, "NOT IMPLEMENT FOR MINERNODE")
}

func (n *MinerNode) CreateExporter() *exporter.Exporter {
	return exporter.NewExporter(n)
}
