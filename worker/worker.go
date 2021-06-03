package worker

import (
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolDevOps/fbc-devops-peer/basenode"
	devops "github.com/NpoolDevOps/fbc-devops-peer/devops"
	exporter "github.com/NpoolDevOps/fbc-devops-peer/exporter"
	"github.com/prometheus/client_golang/prometheus"
)

type WorkerNode struct {
	*basenode.Basenode
}

func NewWorkerNode(config *basenode.BasenodeConfig, devopsClient *devops.DevopsClient) *WorkerNode {
	log.Infof(log.Fields{}, "create %v ndoe", config.NodeConfig.MainRole)
	worker := &WorkerNode{
		basenode.NewBasenode(config, devopsClient),
	}
	return worker
}

func (n *WorkerNode) Describe(ch chan<- *prometheus.Desc) {
	n.BaseMetrics.Describe(ch)
}

func (n *WorkerNode) Collect(ch chan<- prometheus.Metric) {
	n.BaseMetrics.Collect(ch)
}

func (n *WorkerNode) CreateExporter() *exporter.Exporter {
	return exporter.NewExporter(n)
}

func (n *WorkerNode) Banner() {
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
	log.Infof(log.Fields{}, "      WWWWWWOOOOOORRRRRRKKKKK      ")
	log.Infof(log.Fields{}, "IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII")
}
